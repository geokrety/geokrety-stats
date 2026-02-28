#!/usr/bin/env python3
"""
GeoKrety Points System Simulator

Simulates the complete points calculation pipeline for historical moves
based on the documented gamification rules.

Rules implemented:
- Base move points: +3 for first non-owner move
- Waypoint requirement: DROP/SEEN/DIP moves require waypoint to earn points
- Location identification: waypoint (primary) OR coordinates (fallback)
- Owner GK limit: max 10 per owner per user per month
- Waypoint penalty: 50%/25%/0% based on location frequency
- Chain tracking: SEEN (with location) joins chain

Usage:
    python3 simulate_points.py [limit] [host] [user] [password]

Examples:
    python3 simulate_points.py                    # Default: 1000 rows
    python3 simulate_points.py 10000              # 10,000 rows
    python3 simulate_points.py 5000 192.168.130.65 geokrety geokrety
"""

import sys
import psycopg2
from collections import defaultdict
from datetime import datetime
from typing import Dict, List, Optional

# Move type constants
GRAB = 0
DIP = 1
DROP = 2
SEEN = 3
RESCUE = 5

MOVE_TYPE_NAMES = {0: "GRAB", 1: "DIP", 2: "DROP", 3: "SEEN", 5: "RESCUE"}


class MoveRecord:
    """Represents a single move from the database."""

    def __init__(self, db_row):
        """
        Initialize from database tuple with columns:
        id, geokret, lat, lon, country, waypoint, author, move_type, moved_on_datetime
        """
        (self.id, self.geokret, self.lat, self.lon, self.country,
         self.waypoint, self.author, self.move_type, self.moved_on_datetime) = db_row

        # Normalize None/empty values
        self.lat = self._parse_float(self.lat)
        self.lon = self._parse_float(self.lon)
        self.country = self._parse_str(self.country)
        self.waypoint = self._parse_str(self.waypoint)

    @staticmethod
    def _parse_float(val) -> Optional[float]:
        """Handle None, empty strings, and 'None' string values."""
        if val is None or val == '' or (isinstance(val, str) and val.lower() == 'none'):
            return None
        try:
            return float(val)
        except (ValueError, TypeError):
            return None

    @staticmethod
    def _parse_str(val) -> Optional[str]:
        """Handle None, empty strings, and 'None' string values."""
        if val is None or val == '' or (isinstance(val, str) and val.lower() == 'none'):
            return None
        return str(val).strip() if val else None

    def has_location(self) -> bool:
        """Check if move has location data (waypoint OR coordinates)."""
        return self.waypoint is not None or (self.lat is not None and self.lon is not None)

    def location_id(self) -> Optional[str]:
        """Get location identifier: waypoint (primary) OR coordinates (fallback)."""
        if self.waypoint:
            return f"WP:{self.waypoint}"
        elif self.lat is not None and self.lon is not None:
            return f"COORD:{self.lat:.4f}:{self.lon:.4f}"
        return None


class PointsCalculator:
    """Simulates the complete GeoKrety points calculation system."""

    def __init__(self):
        self.user_points = defaultdict(float)
        self.gk_holders = {}
        self.gk_owner = {}
        self.user_moves_per_gk = defaultdict(lambda: defaultdict(list))
        self.locations_by_gk_user_month = defaultdict(set)
        self.move_count = 0
        self.processed_count = 0
        self.skipped_count = 0
        self.move_details = defaultdict(list)  # For debugging

    def process_moves(self, moves: List[MoveRecord]):
        """Process all moves in chronological order."""
        sorted_moves = sorted(moves, key=lambda m: m.moved_on_datetime)

        for move in sorted_moves:
            self.process_move(move)

    def process_move(self, move: MoveRecord):
        """Process a single move through the pipeline."""
        self.move_count += 1

        # 00. Event guard - skip invalid/system moves
        if not move.author or move.move_type is None or move.geokret is None:
            self.skipped_count += 1
            return

        if move.move_type not in [GRAB, DIP, DROP, SEEN, RESCUE]:
            self.skipped_count += 1
            return

        # 01. Initialize GK state on first encounter
        if move.geokret not in self.gk_holders:
            self.gk_holders[move.geokret] = None
            self.gk_owner[move.geokret] = None

        # Update holder state
        if move.move_type in [GRAB, DIP]:
            self.gk_holders[move.geokret] = move.author
        elif move.move_type == DROP:
            self.gk_holders[move.geokret] = None

        # Parse datetime - handle both string and datetime objects
        try:
            if isinstance(move.moved_on_datetime, str):
                dt = datetime.fromisoformat(move.moved_on_datetime.replace('+00', '+00:00'))
            else:
                # Already a datetime object from psycopg2
                dt = move.moved_on_datetime
            month_key = (dt.year, dt.month)
        except Exception as e:
            self.skipped_count += 1
            return

        # 02. Base move points calculation
        base_points = 0
        is_owner = move.author == self.gk_owner.get(move.geokret)
        is_first_move = len(self.user_moves_per_gk[move.author][move.geokret]) == 0

        # Waypoint requirement: DROP/SEEN/DIP need waypoint
        has_waypoint = move.waypoint is not None
        location_requirement_satisfied = (move.move_type in [GRAB, RESCUE] or has_waypoint)

        if is_first_move and not is_owner and location_requirement_satisfied:
            base_points = 3.0

        self.user_moves_per_gk[move.author][move.geokret].append(move)

        if base_points == 0:
            self.skipped_count += 1
            return

        # 04. Waypoint penalty - track location usage
        penalty_factor = 1.0
        if move.has_location():
            location_id = move.location_id()
            # Key: (user, month, location) - NOT including GK
            # This tracks how many GKs this user moved at this location in this month
            key = (move.author, month_key, location_id)

            # Count previous GKs at this location by this user in this month
            prev_gks_count = len(self.locations_by_gk_user_month[key])

            if prev_gks_count == 1:
                penalty_factor = 0.5
            elif prev_gks_count == 2:
                penalty_factor = 0.25
            elif prev_gks_count >= 3:
                # Automatic exclusion - location saturated
                self.skipped_count += 1
                return

            # Add this GK to the set for this location
            self.locations_by_gk_user_month[key].add(move.geokret)

        # Move is valid and will be counted as processed
        self.processed_count += 1

        # Calculate final points
        points = base_points * penalty_factor
        self.user_points[move.author] += points

        move_name = MOVE_TYPE_NAMES.get(move.move_type, 'UNKNOWN')
        self.move_details[move.author].append({
            'gk': move.geokret,
            'type': move_name,
            'base': base_points,
            'penalty': penalty_factor,
            'final': points
        })

    def generate_report(self) -> str:
        """Generate comprehensive simulation report."""
        lines = []
        lines.append("=" * 90)
        lines.append("GeoKrety Points System Simulation Report")
        lines.append("=" * 90)
        lines.append("")

        lines.append(f"Total moves analyzed:  {self.move_count}")
        lines.append(f"Moves processed:       {self.processed_count}")
        lines.append(f"Moves skipped:         {self.skipped_count}")
        lines.append(f"Unique users:          {len(self.user_points)}")
        lines.append(f"Unique GeoKrety items: {len(self.gk_holders)}")
        lines.append("")

        # Top users by points
        if self.user_points:
            lines.append("TOP 25 USERS BY POINTS:")
            lines.append("-" * 90)
            sorted_users = sorted(self.user_points.items(), key=lambda x: x[1], reverse=True)[:25]

            total = sum(self.user_points.values())
            for rank, (user_id, points) in enumerate(sorted_users, 1):
                pct = (points / total * 100) if total > 0 else 0
                move_count = len(self.move_details.get(user_id, []))
                lines.append(f"  {rank:2d}. User {user_id:8d}: {points:9.1f} pts ({pct:5.1f}%) - {move_count} moves")

            lines.append("")
            lines.append(f"Total points awarded: {total:.1f}")
            lines.append(f"Average per user:     {total / len(self.user_points):.1f}")
            lines.append("")

            # Distribution statistics
            points_list = sorted(self.user_points.values(), reverse=True)
            lines.append("DISTRIBUTION ANALYSIS:")
            lines.append("-" * 90)
            lines.append(f"  Median points:       {points_list[len(points_list)//2]:.1f}")
            lines.append(f"  Max points (user):   {points_list[0]:.1f}")
            lines.append(f"  Min points (user):   {points_list[-1]:.1f}")
            lines.append(f"  Range:               {points_list[0] - points_list[-1]:.1f}")
            lines.append("")

        lines.append("=" * 90)
        return "\n".join(lines)


def fetch_moves_from_db(host: str, user: str, password: str, limit: int = 1000) -> List[MoveRecord]:
    """
    Fetch moves from PostgreSQL database.

    Args:
        host: Database host
        user: Database user
        password: Database password
        limit: Number of rows to fetch (default 1000)

    Returns:
        List of MoveRecord objects
    """
    try:
        conn = psycopg2.connect(
            host=host,
            database="geokrety",
            user=user,
            password=password
        )

        cursor = conn.cursor()

        # Query the database - fetch most recent moves
        if limit > 0:
            query = """
                SELECT
                    id, geokret, lat, lon, country, waypoint, author, move_type, moved_on_datetime
                FROM geokrety.gk_moves
                ORDER BY moved_on_datetime DESC
                LIMIT %s
            """
            print(f"Fetching {limit} rows from database...", file=sys.stderr)
            cursor.execute(query, (limit,))
        else:
            # limit <= 0 means fetch all rows
            query = """
                SELECT
                    id, geokret, lat, lon, country, waypoint, author, move_type, moved_on_datetime
                FROM geokrety.gk_moves
                ORDER BY moved_on_datetime DESC
            """
            print(f"Fetching ALL rows from database...", file=sys.stderr)
            cursor.execute(query)

        rows = cursor.fetchall()
        print(f"Fetched {len(rows)} rows from database", file=sys.stderr)

        moves = [MoveRecord(row) for row in rows]

        cursor.close()
        conn.close()

        return moves

    except psycopg2.Error as e:
        print(f"Database connection error: {e}", file=sys.stderr)
        sys.exit(1)


def main():
    """Parse arguments and run simulator."""
    # Parse command line arguments
    limit = 1000  # Default
    host = "192.168.130.65"
    user = "geokrety"
    password = "geokrety"

    if len(sys.argv) > 1:
        try:
            limit = int(sys.argv[1])
        except ValueError:
            print(f"Invalid limit: {sys.argv[1]}", file=sys.stderr)
            sys.exit(1)

    if len(sys.argv) > 2:
        host = sys.argv[2]

    if len(sys.argv) > 3:
        user = sys.argv[3]

    if len(sys.argv) > 4:
        password = sys.argv[4]

    print(f"Configuration:", file=sys.stderr)
    limit_display = "ALL rows" if limit <= 0 else f"{limit} rows"
    print(f"  Limit:   {limit_display}", file=sys.stderr)
    print(f"  Host:    {host}", file=sys.stderr)
    print(f"  User:    {user}", file=sys.stderr)
    print(f"", file=sys.stderr)

    # Fetch from database
    moves = fetch_moves_from_db(host, user, password, limit)

    if not moves:
        print("No moves to process", file=sys.stderr)
        return

    # Process moves
    calc = PointsCalculator()
    calc.process_moves(moves)

    # Print report
    print(calc.generate_report())


if __name__ == '__main__':
    main()
