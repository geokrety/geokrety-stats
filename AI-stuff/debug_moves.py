#!/usr/bin/env python3
"""
Detailed debug output showing exactly which moves are processed and why.
"""

import sys
import psycopg2
from collections import defaultdict
from datetime import datetime

GRAB = 0
DIP = 1
DROP = 2
SEEN = 3
RESCUE = 5

TYPE_NAMES = {0: "GRAB", 1: "DIP", 2: "DROP", 3: "SEEN", 5: "RESCUE"}

class DebugCalculator:
    def __init__(self):
        self.user_points = defaultdict(float)
        self.user_moves_per_gk = defaultdict(lambda: defaultdict(list))
        self.locations_by_gk_user_month = defaultdict(set)
        self.processed_count = 0
        self.skipped_count = 0
        self.skip_reasons = defaultdict(int)
        self.sample_processed = []
        self.sample_skipped = []

    def process_moves(self, moves):
        """Process moves with detailed tracking."""
        for move_tuple in moves:
            move_id, gk, author, move_type, wp, lat, lon, dt_str = move_tuple

            skip_reason = None

            # Event guard
            if not author or move_type is None or gk is None:
                skip_reason = "no author/type/gk"
            elif move_type not in [GRAB, DIP, DROP, SEEN, RESCUE]:
                skip_reason = f"invalid type {move_type}"
            else:
                # Parse datetime
                try:
                    if isinstance(dt_str, str):
                        dt = datetime.fromisoformat(dt_str.replace('+00', '+00:00'))
                    else:
                        dt = dt_str
                    month_key = (dt.year, dt.month)
                except:
                    skip_reason = "bad datetime"

                if not skip_reason:
                    # Check first move
                    is_first = len(self.user_moves_per_gk[author][gk]) == 0

                    # Waypoint requirement
                    has_wp = (wp is not None)
                    loc_req_ok = move_type in [GRAB, RESCUE] or has_wp

                    if not is_first:
                        skip_reason = "not first move"
                    elif not loc_req_ok:
                        type_name = TYPE_NAMES.get(move_type, "?")
                        skip_reason = f"{type_name} no waypoint"
                    else:
                        # Valid for points
                        self.user_moves_per_gk[author][gk].append(move_id)

                        # Apply penalty
                        base_points = 3.0
                        penalty = 1.0

                        if wp or (lat is not None and lon is not None):
                            if wp:
                                loc_id = f"WP:{wp}"
                            else:
                                loc_id = f"COORD:{lat:.4f}:{lon:.4f}"

                            key = (author, month_key, loc_id)  # Key WITHOUT GK
                            prev_gks_count = len(self.locations_by_gk_user_month[key])

                            if prev_gks_count == 1:
                                penalty = 0.5
                            elif prev_gks_count == 2:
                                penalty = 0.25
                            elif prev_gks_count >= 3:
                                skip_reason = f"location saturated (4th+ GK at {loc_id})"
                                penalty = 0

                            self.locations_by_gk_user_month[key].add(gk)

                        if penalty > 0:
                            points = base_points * penalty
                            self.user_points[author] += points
                            self.processed_count += 1

                            type_name = TYPE_NAMES.get(move_type, "?")
                            if len(self.sample_processed) < 10:
                                self.sample_processed.append({
                                    'move_id': move_id, 'gk': gk, 'user': author, 'type': type_name,
                                    'base': base_points, 'penalty': penalty, 'points': points
                                })
                        else:
                            self.skipped_count += 1
                            self.skip_reasons[skip_reason] += 1

            if skip_reason:
                self.skipped_count += 1
                self.skip_reasons[skip_reason] += 1
                if len(self.sample_skipped) < 15:
                    type_name = TYPE_NAMES.get(move_type, "?")
                    self.sample_skipped.append({
                        'move_id': move_id, 'gk': gk, 'user': author, 'type': type_name, 'reason': skip_reason
                    })


conn = psycopg2.connect(host="192.168.130.65", database="geokrety", user="geokrety", password="geokrety")
cursor = conn.cursor()

print("=== DEBUG: PROCESSING 1000 MOST RECENT MOVES ===\n")

cursor.execute("""
    SELECT id, geokret, author, move_type, waypoint, lat, lon, moved_on_datetime
    FROM geokrety.gk_moves
    ORDER BY moved_on_datetime DESC
    LIMIT 1000
""")

moves = cursor.fetchall()

calc = DebugCalculator()
calc.process_moves(moves)

print(f"PROCESSED: {calc.processed_count} / SKIPPED: {calc.skipped_count}\n")

print("SAMPLE PROCESSED MOVES:")
for m in calc.sample_processed:
    print(f"  Move {m['move_id']:8d}: {m['type']:8s} User {m['user']:8d} GK {m['gk']:8d} -> {m['points']:6.2f} pts (base {m['base']:.1f} * penalty {m['penalty']})")

print(f"\nSKIP REASONS ({calc.skipped_count} total):")
for reason, count in sorted(calc.skip_reasons.items(), key=lambda x: -x[1])[:15]:
    pct = (count / (calc.processed_count + calc.skipped_count) * 100)
    # Truncate long reasons
    reason_display = reason if len(reason) <= 50 else reason[:47] + "..."
    print(f"  {count:4d} ({pct:5.1f}%) - {reason_display}")

print(f"\nAverage GK per processed move: {len(set(m[1] for m in moves)) / calc.processed_count:.1f}" if calc.processed_count > 0 else "N/A")

cursor.close()
conn.close()
