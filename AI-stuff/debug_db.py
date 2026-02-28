#!/usr/bin/env python3
"""Debug script to inspect database columns and sample data."""

import psycopg2

conn = psycopg2.connect(
    host="192.168.130.65",
    database="geokrety",
    user="geokrety",
    password="geokrety"
)

cursor = conn.cursor()

# Get column names
cursor.execute("""
    SELECT column_name, data_type
    FROM information_schema.columns
    WHERE table_name='gk_moves' AND table_schema='geokrety'
    ORDER BY ordinal_position
""")

print("=== GK_MOVES Table Columns ===")
columns = cursor.fetchall()
for col, dtype in columns:
    print(f"  {col:25s} {dtype}")

print("\n=== Sample Rows (First 5) ===")

cursor.execute("""
    SELECT id, geokret, lat, lon, country, waypoint, author, move_type, moved_on_datetime
    FROM geokrety.gk_moves
    ORDER BY moved_on_datetime DESC
    LIMIT 5
""")

rows = cursor.fetchall()
for i, row in enumerate(rows, 1):
    print(f"\nRow {i}:")
    print(f"  id:                 {row[0]} (type: {type(row[0]).__name__})")
    print(f"  geokret:            {row[1]} (type: {type(row[1]).__name__})")
    print(f"  lat:                {row[2]} (type: {type(row[2]).__name__})")
    print(f"  lon:                {row[3]} (type: {type(row[3]).__name__})")
    print(f"  country:            {row[4]} (type: {type(row[4]).__name__})")
    print(f"  waypoint:           {row[5]} (type: {type(row[5]).__name__})")
    print(f"  author:             {row[6]} (type: {type(row[6]).__name__})")
    print(f"  move_type:          {row[7]} (type: {type(row[7]).__name__})")
    print(f"  moved_on_datetime:  {row[8]} (type: {type(row[8]).__name__})")

cursor.close()
conn.close()
