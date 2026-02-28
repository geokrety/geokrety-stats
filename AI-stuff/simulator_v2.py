#!/usr/bin/env python3
"""
GeoKrety Points System Simulator v2 - Complete Implementation

MOVE TYPE CONSTANTS (CORRECTED):
  0 = DROP      (place item in cache)
  1 = GRAB      (take item from cache)
  2 = COMMENT   (not scoreable)
  3 = SEEN      (observed in cache)
  4 = ARCHIVED  (end of life)
  5 = DIP       (dipped in place)

Rules Implemented:
  Module 00: Event Guard (auth, scoreable types)
  Module 01: Context Loader (state loading)
  Module 02: Base Move Points (3.0 per first move)
  Module 03: Owner GK Limit (10 distinct per owner)
  Module 04: Waypoint Penalty (location saturation)
"""

import psycopg2
import sys
from collections import defaultdict
from datetime import datetime

# ============================================================================
# CONSTANTS
# ============================================================================

MOVE_DROP = 0
MOVE_GRAB = 1
MOVE_COMMENT = 2
MOVE_SEEN = 3
MOVE_ARCHIVED = 4
MOVE_DIP = 5

NAMES = {0: "DROP", 1: "GRAB", 2: "COMMENT", 3: "SEEN", 4: "ARCHIVED", 5: "DIP"}

BASE_POINTS = 3.0
OWNER_LIMIT = 10
WAYPOINT_PENALTIES = {0: 1.0, 1: 0.5, 2: 0.25}  # By index (0=100%, 1=50%, 2=25%)


class Move:
    """Single move record."""
    def __init__(self, row):
        self.id = row[0]
        self.gk_id = row[1]
        self.lat = row[2]
        self.lon = row[3]
        self.country = row[4]
        self.waypoint = row[5]
        self.user_id = row[6]
        self.move_type = row[7]
        self.timestamp = row[8]

        # Parse month
        try:
            if isinstance(self.timestamp, str):
                dt = datetime.fromisoformat(self.timestamp.replace('+00', '+00:00'))
            else:
                dt = self.timestamp
            self.month_key = (dt.year, dt.month)
        except:
            self.month_key = None

    def location(self):
        """Get location ID (waypoint > coordinates > None)."""
        if self.waypoint:
            return ("W", self.waypoint)
        if self.lat is not None and self.lon is not None:
            return ("C", (self.lat, self.lon))
        return None


class Simulator:
    def __init__(self):
        self.total_moves = 0
        self.processed = 0
        self.skipped = 0

        self.points_map = defaultdict(float)
        self.skip_reasons = defaultdict(int)
        self.move_type_counts = defaultdict(int)

        # State tracking
        self.gk_holders = {}  # gk -> current holder or None
        self.gk_owners = {}   # gk -> owner
        self.user_gk_count_per_owner = defaultdict(lambda: defaultdict(set))  # user -> owner -> {gk_ids}
        self.user_count_at_location_month = defaultdict(int)  # (user, month, location) -> count
        self.user_gk_history_on_gk = defaultdict(lambda: defaultdict(set))  # user -> gk -> {move types}

    def process(self, moves):
        """Process all moves."""
        for move in moves:
            self.total_moves += 1
            self._process_one(move)

    def _process_one(self, m):
        """Process single move."""
        # Count move types
        self.move_type_counts[m.move_type] += 1

        # MODULE 00: EVENT GUARD
        if m.user_id is None:
            self.skipped += 1
            self.skip_reasons["anonymous"] += 1
            return

        if m.move_type not in {MOVE_DROP, MOVE_GRAB, MOVE_SEEN, MOVE_ARCHIVED, MOVE_DIP}:
            self.skipped += 1
            self.skip_reasons["invalid_type"] += 1
            return

        if m.month_key is None:
            self.skipped += 1
            self.skip_reasons["bad_timestamp"] += 1
            return

        # MODULE 01: CONTEXT
        if m.gk_id not in self.gk_holders:
            self.gk_holders[m.gk_id] = None
            self.gk_owners[m.gk_id] = None

        # Update holder
        if m.move_type in {MOVE_GRAB, MOVE_DIP}:
            self.gk_holders[m.gk_id] = m.user_id
        elif m.move_type == MOVE_DROP:
            self.gk_holders[m.gk_id] = None

        # MODULE 02: BASE POINTS
        # ARCHIVED and DIP don't give base points
        if m.move_type in {MOVE_ARCHIVED, MOVE_DIP}:
            self.skipped += 1
            self.skip_reasons[NAMES[m.move_type] + "_no_base"] += 1
            return

        # Waypoint required for DROP, SEEN, DIP (dip doesn't reach here)
        if m.move_type in {MOVE_DROP, MOVE_SEEN}:
            if not m.location():
                self.skipped += 1
                self.skip_reasons["no_location"] += 1
                return

        # Check first move
        is_owner = m.user_id == self.gk_owners.get(m.gk_id)
        is_first_move = m.move_type not in self.user_gk_history_on_gk[m.user_id][m.gk_id]

        # Track history
        self.user_gk_history_on_gk[m.user_id][m.gk_id].add(m.move_type)

        if not is_first_move:
            self.skipped += 1
            self.skip_reasons["not_first_move"] += 1
            return

        if is_owner:
            self.skipped += 1
            self.skip_reasons["owner_move"] += 1
            return

        base_pts = BASE_POINTS

        # MODULE 03: OWNER LIMIT
        owner = self.gk_owners.get(m.gk_id)
        if owner is not None and not is_owner:
            owner_gks = self.user_gk_count_per_owner[m.user_id][owner]
            if m.gk_id not in owner_gks:
                if len(owner_gks) >= OWNER_LIMIT:
                    self.skipped += 1
                    self.skip_reasons["owner_limit"] += 1
                    return
                owner_gks.add(m.gk_id)

        # MODULE 04: WAYPOINT PENALTY
        penalty = 1.0
        if m.location():
            loc_key = (m.user_id, m.month_key, m.location())
            prev_count = self.user_count_at_location_month[loc_key]

            if prev_count >= 3:
                penalty = 0.0
                self.skipped += 1
                self.skip_reasons["location_sat"] += 1
                return
            else:
                penalty = WAYPOINT_PENALTIES.get(prev_count, 0.0)

            self.user_count_at_location_month[loc_key] += 1

        # AWARD POINTS
        points = base_pts * penalty
        self.points_map[m.user_id] += points
        self.processed += 1

    def report(self, usernames):
        """Generate report."""
        lines = ["=" * 120]
        lines.append("GeoKrety Points Simulation Report (v2)")
        lines.append("=" * 120)
        lines.append("")
        lines.append(f"Total moves:         {self.total_moves:,}")
        lines.append(f"Processed:           {self.processed:,}")
        lines.append(f"Skipped:             {self.skipped:,}")
        lines.append(f"Unique users:        {len(self.points_map):,}")
        lines.append("")

        lines.append("MOVE TYPE DISTRIBUTION:")
        lines.append("-" * 120)
        for move_type in sorted(self.move_type_counts.keys()):
            count = self.move_type_counts[move_type]
            pct = 100.0 * count / self.total_moves
            lines.append(f"  {NAMES.get(move_type, f'TYPE_{move_type}'):10s}: {count:8,} ({pct:6.2f}%)")
        lines.append("")

        lines.append("SKIP REASONS:")
        lines.append("-" * 120)
        for reason in sorted(self.skip_reasons.keys(), key=lambda x: -self.skip_reasons[x]):
            count = self.skip_reasons[reason]
            pct = 100.0 * count / self.skipped if self.skipped > 0 else 0
            lines.append(f"  {reason:25s}: {count:8,} ({pct:6.2f}%)")
        lines.append("")

        lines.append("TOP 25 USERS:")
        lines.append("-" * 120)
        sorted_users = sorted(self.points_map.items(), key=lambda x: -x[1])[:25]
        total = sum(self.points_map.values())

        for rank, (uid, pts) in enumerate(sorted_users, 1):
            pct = 100.0 * pts / total if total > 0 else 0
            name = usernames.get(uid, f"User{uid}")
            lines.append(f"  {rank:2d}. {name:25s}  ID{uid:8d}: {pts:10.1f} pts ({pct:5.1f}%)")

        lines.append("")
        lines.append(f"Total points:        {total:,.1f}")
        lines.append(f"Avg per user:        {total / len(self.points_map):.1f}" if self.points_map else "N/A")
        lines.append("")
        lines.append("=" * 120)
        return "\n".join(lines)


def get_moves(host, user, pw, limit):
    """Fetch moves from DB."""
    conn = psycopg2.connect(host=host, database="geokrety", user=user, password=pw)
    cur = conn.cursor()

    if limit > 0:
        print(f"Fetching {limit:,} moves...", file=sys.stderr)
        cur.execute(
            "SELECT id, geokret, lat, lon, country, waypoint, author, move_type, moved_on_datetime "
            "FROM geokrety.gk_moves ORDER BY moved_on_datetime ASC LIMIT %s",
            (limit,)
        )
    else:
        print(f"Fetching ALL moves...", file=sys.stderr)
        cur.execute(
            "SELECT id, geokret, lat, lon, country, waypoint, author, move_type, moved_on_datetime "
            "FROM geokrety.gk_moves ORDER BY moved_on_datetime ASC"
        )

    rows = cur.fetchall()
    print(f"Loaded {len(rows):,} moves", file=sys.stderr)
    cur.close()
    conn.close()

    return [Move(r) for r in rows]


def get_usernames(host, user, pw):
    """Fetch usernames."""
    try:
        conn = psycopg2.connect(host=host, database="geokrety", user=user, password=pw)
        cur = conn.cursor()
        cur.execute("SELECT id, username FROM geokrety.gk_users WHERE username IS NOT NULL LIMIT 50000")
        names = {uid: uname for uid, uname in cur.fetchall()}
        cur.close()
        conn.close()
        print(f"Loaded {len(names):,} usernames", file=sys.stderr)
        return names
    except Exception as e:
        print(f"Warning: UserNames failed: {e}", file=sys.stderr)
        return {}


def main():
    limit = 1000
    if len(sys.argv) > 1:
        limit = int(sys.argv[1])

    host = sys.argv[2] if len(sys.argv) > 2 else "192.168.130.65"
    user = sys.argv[3] if len(sys.argv) > 3 else "geokrety"
    pw = sys.argv[4] if len(sys.argv) > 4 else "geokrety"

    print(f"Configuration:", file=sys.stderr)
    print(f"  Limit: {'ALL' if limit <= 0 else f'{limit:,}'}", file=sys.stderr)
    print(f"  Fetching data...", file=sys.stderr)

    moves = get_moves(host, user, pw, limit)
    names = get_usernames(host, user, pw)

    sim = Simulator()
    sim.process(moves)

    print(sim.report(names))


if __name__ == "__main__":
    main()
