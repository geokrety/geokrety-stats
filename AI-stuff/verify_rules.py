#!/usr/bin/env python3
"""
Data analysis for verification.
"""

import psycopg2

conn = psycopg2.connect(host="192.168.130.65", database="geokrety", user="geokrety", password="geokrety")
cursor = conn.cursor()

print("=== ANALYZING MOST RECENT 1000 MOVES ===\n")

cursor.execute("SELECT id, geokret, author, move_type, waypoint, lat, lon FROM geokrety.gk_moves ORDER BY moved_on_datetime DESC LIMIT 1000")
moves = cursor.fetchall()

# Analyze
type_map = {0: "GRAB", 1: "DIP", 2: "DROP", 3: "SEEN", 5: "RESCUE"}
type_counts = {}
no_author_count = 0
location_viol = {}

for move_id, gk, author, mtype, wp, lat, lon in moves:
    type_name = type_map.get(mtype, f"TYPE_{mtype}")

    if type_name not in type_counts:
        type_counts[type_name] = 0
    type_counts[type_name] += 1

    if author is None:
        no_author_count += 1

    # Check waypoint requirement
    if mtype in [1, 2, 3]:  # DIP, DROP, SEEN
        if wp is None and lat is None:
            if type_name not in location_viol:
                location_viol[type_name] = 0
            location_viol[type_name] += 1

print("Move type distribution:")
for type_name in sorted(type_counts.keys()):
    print(f"  {type_name:8s}: {type_counts[type_name]:4d}")

print(f"\nMoves with NULL author: {no_author_count}")

print("\nLocation requirement violations (no waypoint AND no coords):")
for type_name in sorted(location_viol.keys()):
    print(f"  {type_name:8s}: {location_viol[type_name]:4d}")

print(f"\nTotal violations: {sum(location_viol.values())}")

cursor.close()
conn.close()
