#!/usr/bin/env python3
"""
Trace specific moves through simulator to verify penalty calculation.
"""

import psycopg2
from datetime import datetime
from collections import defaultdict

RESCUE = 5
TYPE_NAMES = {5: "RESCUE"}

conn = psycopg2.connect(host="192.168.130.65", database="geokrety", user="geokrety", password="geokrety")
cursor = conn.cursor()

print("=== TRACING PENALTY CALCULATION ===\n")

# Get the specific moves
cursor.execute("""
    SELECT id, geokret, author, move_type, waypoint, lat, lon, moved_on_datetime
    FROM geokrety.gk_moves
    WHERE author = 36628
    AND waypoint = 'VI29452'
    AND EXTRACT(YEAR FROM moved_on_datetime) = 2023
    AND EXTRACT(MONTH FROM moved_on_datetime) = 6
    ORDER BY moved_on_datetime ASC
    LIMIT 10
""")

moves = cursor.fetchall()
print(f"Processing {len(moves)} moves by user 36628 at waypoint VI29452 in June 2023:\n")

# Simulate
user_points = defaultdict(float)
user_moves_per_gk = defaultdict(lambda: defaultdict(list))
locations_by_gk_user_month = defaultdict(set)

for row_num, (move_id, gk, author, move_type, wp, lat, lon, dt) in enumerate(moves, 1):
    # Check first move
    is_first = len(user_moves_per_gk[author][gk]) == 0

    if is_first:
        user_moves_per_gk[author][gk].append(move_id)

        # Calculate penalty
        dt_obj = dt
        month_key = (dt_obj.year, dt_obj.month)

        base_points = 3.0
        penalty = 1.0

        # Build location key - CORRECT: (user, month, location) NOT including GK
        if wp:
            loc_id = f"WP:{wp}"
        elif lat and lon:
            loc_id = f"COORD:{lat:.4f}:{lon:.4f}"
        else:
            loc_id = None

        if loc_id:
            key = (author, month_key, loc_id)  # Key without GK
            prev_gks_count = len(locations_by_gk_user_month[key])

            if prev_gks_count == 0:
                penalty = 1.0
            elif prev_gks_count == 1:
                penalty = 0.5
            elif prev_gks_count == 2:
                penalty = 0.25
            elif prev_gks_count >= 3:
                penalty = 0

            locations_by_gk_user_month[key].add(gk)

        points = base_points * penalty
        user_points[author] += points

        penalty_pct = int(penalty * 100)
        print(f"  Move {row_num:2d} (ID {move_id:8d}): GK {gk:8d} -> First move! Base 3.0 * {penalty} = {points:.2f} pts ({penalty_pct}%)")
    else:
        print(f"  Move {row_num:2d} (ID {move_id:8d}): GK {gk:8d} -> Not first move, skip (0 pts)")

cursor.close()
conn.close()

print(f"\nTotal points for user 36628: {user_points[36628]:.2f}")
print(f"Expected: 3.0 + 1.5 + 0.75 + 0 + 0 + 0 + 0 + 0 + 0 + 0 = 5.25")
