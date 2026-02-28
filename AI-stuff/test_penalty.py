#!/usr/bin/env python3
"""
Specific test to verify waypoint penalty (100% -> 50% -> 25% -> 0%).
"""

import psycopg2
from datetime import datetime

conn = psycopg2.connect(host="192.168.130.65", database="geokrety", user="geokrety", password="geokrety")
cursor = conn.cursor()

print("=== TESTING WAYPOINT PENALTY RULE ===\n")

print("Rule: User moving multiple GKs at same waypoint per month")
print("      1st GK: 100% (3.0 pts)")
print("      2nd GK: 50%  (1.5 pts)")
print("      3rd GK: 25%  (0.75 pts)")
print("      4th+:   0%   (0 pts)\n")

# Find a user who moved multiple GKs at the same waypoint in same month
cursor.execute("""
    SELECT author, waypoint,
           EXTRACT(YEAR FROM moved_on_datetime) as year,
           EXTRACT(MONTH FROM moved_on_datetime) as month,
           COUNT(DISTINCT geokret) as gk_count,
           COUNT(*) as move_count
    FROM geokrety.gk_moves
    WHERE waypoint IS NOT NULL
    GROUP BY author, waypoint, year, month
    HAVING COUNT(DISTINCT geokret) >= 2
    ORDER BY COUNT(DISTINCT geokret) DESC,
             EXTRACT(YEAR FROM moved_on_datetime) DESC,
             EXTRACT(MONTH FROM moved_on_datetime) DESC
    LIMIT 5
""")

results = cursor.fetchall()
print(f"Found {len(results)} cases of users moving multiple GKs at same waypoint in same month:\n")

for user, wp, year, month, gk_count, move_count in results:
    print(f"User {user:8d} at waypoint {wp:10s} in {int(year)}-{int(month):02d}: {int(gk_count)} GKs ({int(move_count)} moves)")

    # Get details
    cursor.execute("""
        SELECT geokret, move_type, moved_on_datetime
        FROM geokrety.gk_moves
        WHERE author = %s AND waypoint = %s
        AND EXTRACT(YEAR FROM moved_on_datetime) = %s
        AND EXTRACT(MONTH FROM moved_on_datetime) = %s
        ORDER BY moved_on_datetime ASC
        LIMIT 10
    """, (user, wp, year, month))

    type_map = {0: "GRAB", 1: "DIP", 2: "DROP", 3: "SEEN", 5: "RESCUE"}

    for row_num, (gk, move_type, dt) in enumerate(cursor.fetchall(), 1):
        type_name = type_map.get(move_type, "?")
        expected_penalty = {1: "100%", 2: "50%", 3: "25%", 4: "0%"}.get(min(row_num, 4), "N/A")
        print(f"    Move {row_num}: GK {gk:8d} {type_name:8s} on {dt.date()} -> expect {expected_penalty}")
    print()

cursor.close()
conn.close()
