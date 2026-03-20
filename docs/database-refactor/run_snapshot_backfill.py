#!/usr/bin/env python3

# pyright: reportMissingTypeStubs=false

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import os
import re
import sys
import threading
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from dataclasses import dataclass, field
from typing import Iterable, Iterator, Optional, Union
from threading import Lock

import psycopg2

try:
    from rich.console import Console, Group
    from rich.live import Live
    from rich.panel import Panel
    from rich.table import Table

    HAS_RICH = True
except ImportError:
    HAS_RICH = False

# Try to import colorama for cross-platform colored output
try:
    from colorama import Fore, Back, Style, init
    init(autoreset=True)
    HAS_COLOR = True
except ImportError:
    HAS_COLOR = False
    # Fallback no-op color codes
    class Fore:
        GREEN = ""
        YELLOW = ""
        RED = ""
        CYAN = ""
        WHITE = ""
    class Back:
        pass
    class Style:
        RESET_ALL = ""
        BRIGHT = ""


UTC = dt.timezone.utc
DEFAULT_START = dt.datetime(2007, 10, 1, tzinfo=UTC)
RESUME_VERSION = 1
RUNNER_JOB_NAME = "run_snapshot_backfill_step"
PRE_FULL_ONCE_PHASES = (
    "fn_snapshot_entity_counters",
    "fn_snapshot_daily_entity_counts",
)
POST_FULL_ONCE_PHASES = (
    "fn_snapshot_gk_country_history",
    "fn_snapshot_first_finder_events",
    "fn_snapshot_gk_milestone_events",
)
REPLICA_ROLE_PHASES: tuple[str, ...] = (
    "fn_snapshot_daily_entity_counts",
    "fn_snapshot_gk_country_history",
    "fn_snapshot_first_finder_events",
    "fn_snapshot_gk_milestone_events",
)
# Phases that must complete before snapshot phases (writes previous_move_id)
SEQUENTIAL_MONTHLY_PHASES: tuple[str, ...] = (
    "fn_backfill_heavy_previous_move_id_all",
)
# Phases that are independent of each other and run in parallel by default
PARALLEL_MONTHLY_PHASES: tuple[str, ...] = (
    "fn_seed_daily_activity",
    "fn_snapshot_daily_country_stats",
    "fn_snapshot_user_country_stats",
    "fn_snapshot_gk_country_stats",
    "fn_snapshot_relationship_tables",
    "fn_snapshot_hourly_activity",
    "fn_snapshot_country_pair_flows",
)
# Combined tuple (used when --no-parallel)
MONTHLY_PHASES: tuple[str, ...] = SEQUENTIAL_MONTHLY_PHASES + PARALLEL_MONTHLY_PHASES

# Unicode symbols
EMOJI = {
    "pending": "⏳",
    "running": "⚙️ ",
    "done": "✅",
    "error": "❌",
    "info": "ℹ️",
    "clock": "🕐",
    "stats": "📊",
}


class TableDisplay:
    """Maintains a clean table display in terminal with in-place updates."""

    def __init__(self):
        self.header_printed = False
        self.last_row_length = 0
        self.last_line_was_incomplete = False
        self.status_line_active = False
        self.output_lock = Lock()
        self.output_lock = Lock()  # Synchronize stdout access

    def print_header(self):
        """Print table header."""
        if self.header_printed:
            return
        with self.output_lock:
            header = (
                f"  {Fore.WHITE}{Style.BRIGHT}#  │"
                f" Phase (50 chars)                          │"
                f" Slice (25 chars)         │"
                f" Elapsed  │"
                f" ETA (UTC){Style.RESET_ALL}"
            )
            print(header)
            print(f"  {Fore.WHITE}{'─' * 115}{Style.RESET_ALL}")
            self.header_printed = True

    def print_row(self, idx: int, total: int, phase: str, slice_label: str,
                  elapsed: float, eta_dt: Optional[dt.datetime],
                  global_eta_dt: Optional[dt.datetime] = None,
                  is_current: bool = False, overwrite: bool = False) -> None:
        """Print or update a progress row.

        Args:
            overwrite: If True, use \\r to overwrite the line (for running tasks).
        """
        self.print_header()

        phase_display = phase[:48].ljust(48)
        slice_display = slice_label[:23].ljust(23)
        elapsed_display = format_duration(elapsed)
        eta_display = (
            eta_dt.strftime("%Y-%m-%d %H:%M:%S")
            if eta_dt else "—" * 19
        )

        # All rows now have emoji: running (⚙️) or done (✅)
        status_icon = EMOJI["running"] if is_current else EMOJI["done"]
        color = Fore.CYAN if is_current else Fore.GREEN

        row = (
            f"{status_icon} {color}{idx:3}│"
            f" {phase_display}│"
            f" {slice_display}│"
            f" {elapsed_display}│"
            f" {eta_display}{Style.RESET_ALL}"
        )

        with self.output_lock:
            if overwrite and self.last_line_was_incomplete:
                # Overwrite previous task line in-place (no newline)
                sys.stdout.write(f"\r  {row}\033[K")
                sys.stdout.flush()
            else:
                # Print normal line with newline
                print(f"  {row}")
                self.last_line_was_incomplete = overwrite

    def print_global_status(self, completed: int, total: int,
                           global_eta_dt: Optional[dt.datetime],
                           overall_elapsed: float) -> None:
        """Print global progress status line with proper line handling."""
        with self.output_lock:
            # Clear any incomplete line first
            if self.last_line_was_incomplete:
                sys.stdout.write("\n")
                sys.stdout.flush()
                self.last_line_was_incomplete = False

            if global_eta_dt is None:
                eta_str = "calculating..."
            else:
                remaining = global_eta_dt - dt.datetime.now(UTC)
                remaining_secs = max(0, remaining.total_seconds())
                eta_str = (
                    f"{global_eta_dt.strftime('%Y-%m-%d %H:%M:%S')} | "
                    f"~{format_duration(remaining_secs)} remaining"
                )

            progress_pct = int((completed / total * 100)) if total > 0 else 0
            status = (
                f"{EMOJI['clock']} Global ETA: {eta_str} | "
                f"Progress: {completed}/{total} ({progress_pct}%) | "
                f"Elapsed: {format_duration(overall_elapsed)}"
            )
            # Write status line with newline at the end
            sys.stdout.write(f"\r{status}\033[K\n")
            sys.stdout.flush()
            self.status_line_active = True

    def finalize_running_line(self) -> None:
        """Move to new line if we were overwriting."""
        with self.output_lock:
            if self.last_line_was_incomplete:
                sys.stdout.write("\n")
                sys.stdout.flush()
                self.last_line_was_incomplete = False

    def print_summary(self, total_time: float, step_count: int) -> None:
        """Print final summary."""
        self.finalize_running_line()
        with self.output_lock:
            print(f"\n  {Fore.WHITE}{'═' * 115}{Style.RESET_ALL}")
            throughput = step_count / (total_time / 3600.0) if total_time > 0 else 0
            print(
                f"  {EMOJI['stats']} {Fore.GREEN}{Style.BRIGHT}"
                f"Completed {Fore.CYAN}{step_count}{Fore.GREEN} steps in "
                f"{format_duration(total_time)}"
                f" ({throughput:.1f} steps/hour){Style.RESET_ALL}"
            )


class EventLogger:
    """Writes real-time execution events to one log file."""

    def __init__(self, file_path: str):
        self.file_path = file_path
        self._lock = Lock()
        os.makedirs(os.path.dirname(file_path) or ".", exist_ok=True)
        # Line-buffered so events are visible while the process is running.
        self._fh = open(file_path, "a", encoding="utf-8", buffering=1)

    def log(self, message: str) -> None:
        ts = dt.datetime.now(UTC).strftime("%Y-%m-%d %H:%M:%S+00")
        line = f"[{ts}] {message}"
        with self._lock:
            self._fh.write(line + "\n")

    def close(self) -> None:
        with self._lock:
            self._fh.close()


class RichDashboard:
    """Live terminal dashboard powered by rich."""

    def __init__(
        self,
        total_steps: int,
        start: dt.datetime,
        end: dt.datetime,
        mode_label: str,
        log_file: str,
    ):
        self.total_steps = total_steps
        self.start = start
        self.end = end
        self.mode_label = mode_label
        self.log_file = log_file
        self.console = Console()
        self.live = Live(
            self._placeholder(),
            console=self.console,
            refresh_per_second=4,
            transient=False,
        )

    def _placeholder(self):
        return Panel("Starting dashboard...", title="Snapshot Backfill")

    def start_live(self) -> None:
        self.live.start()

    def stop_live(self) -> None:
        self.live.stop()

    def render(self, state: "ProgressState") -> None:
        self.live.update(self._build_renderable(state), refresh=True)

    def _build_renderable(self, state: "ProgressState"):
        elapsed = max(0.0, time.monotonic() - state.overall_started_at)
        completed = state.completed_steps
        total = state.total_steps
        pct = (completed / total * 100.0) if total else 0.0

        summary = Table.grid(padding=(0, 1))
        summary.add_column(style="bold cyan", justify="right")
        summary.add_column(style="white")
        summary.add_row("Range", f"{format_ts(self.start)} -> {format_ts(self.end)}")
        summary.add_row("Mode", self.mode_label)
        summary.add_row("Progress", f"{completed}/{total} ({pct:.1f}%)")
        summary.add_row("Elapsed", format_duration(elapsed))
        summary.add_row("Current", state.current_phase or "-")
        summary.add_row("Slice", state.current_slice or "-")
        summary.add_row("Log file", self.log_file)

        active = Table(show_header=True, header_style="bold magenta")
        active.add_column("Phase", overflow="fold")
        active.add_column("Elapsed", justify="right")
        with state.parallel_lock:
            running = sorted(state.parallel_tasks_started_at.items())
        now = time.monotonic()
        if running:
            for phase_name, started in running:
                active.add_row(
                    short_phase_name(phase_name),
                    format_duration(max(0.0, now - started)),
                )
        else:
            active.add_row("-", "-")

        events = Table(show_header=True, header_style="bold green")
        events.add_column("Recent events", overflow="fold")
        with state.events_lock:
            recent = list(state.recent_events[-12:])
        if recent:
            for line in recent:
                events.add_row(line)
        else:
            events.add_row("No events yet")

        return Group(
            Panel(summary, title="Snapshot Backfill", border_style="cyan"),
            Panel(active, title="Active Parallel Phases", border_style="magenta"),
            Panel(events, title="Live Event Stream", border_style="green"),
        )


@dataclass(frozen=True)
class Slice:
    start: dt.datetime
    end: dt.datetime

    @property
    def label(self) -> str:
        return f"{format_ts(self.start)} -> {format_ts(self.end)}"

    @property
    def progress_label(self) -> str:
        return (
            f"{self.start.strftime('%Y-%m-%d')}"
            f"..{self.end.strftime('%Y-%m-%d')}"
        )


@dataclass
class ProgressState:
    total_steps: int
    completed_steps: int = 0
    current_phase: str = ""
    current_slice: str = ""
    current_started_at: float = 0.0
    overall_started_at: float = 0.0
    stop: bool = False
    table_display: TableDisplay = field(default_factory=TableDisplay)
    steps_timings: list[float] = field(default_factory=list)
    parallel_tasks_started_at: dict[str, float] = field(default_factory=dict)
    parallel_lock: Lock = field(default_factory=Lock)
    recent_events: list[str] = field(default_factory=list)
    events_lock: Lock = field(default_factory=Lock)
    dashboard: Optional[RichDashboard] = None


def format_ts(value: dt.datetime) -> str:
    return value.astimezone(UTC).strftime("%Y-%m-%d %H:%M:%S+00")


def format_duration(seconds: float) -> str:
    whole = max(0, int(seconds))
    hours, remainder = divmod(whole, 3600)
    minutes, secs = divmod(remainder, 60)
    return f"{hours:02d}:{minutes:02d}:{secs:02d}"


def stable_json_dumps(value: object) -> str:
    return json.dumps(value, sort_keys=True, separators=(",", ":"))


def short_phase_name(phase: str) -> str:
    return phase.replace("fn_snapshot_", "").replace("fn_backfill_", "")


def phase_kind(phase: str, slice_period: Optional[Slice]) -> str:
    if slice_period is not None:
        return "slice"
    if phase in PRE_FULL_ONCE_PHASES:
        return "pre_full"
    if phase in POST_FULL_ONCE_PHASES:
        return "post_full"
    return "full"


def advisory_lock_keys(label: str) -> tuple[int, int]:
    digest = hashlib.sha256(label.encode("utf-8")).digest()
    return (
        int.from_bytes(digest[:4], byteorder="big", signed=True),
        int.from_bytes(digest[4:8], byteorder="big", signed=True),
    )


def build_run_identity(
    start: dt.datetime,
    end: dt.datetime,
    source_start: dt.datetime,
    source_end: dt.datetime,
    batch_size: int,
    parallel: bool,
    use_replica_role: bool,
    skip_entity_counters: bool,
) -> dict[str, object]:
    requested_start = format_ts(start)
    requested_end = format_ts(end)
    effective_source_start = format_ts(source_start)
    effective_source_end = format_ts(source_end)
    payload: dict[str, object] = {
        "resume_version": RESUME_VERSION,
        "requested_start": requested_start,
        "requested_end": requested_end,
        "source_start": effective_source_start,
        "source_end": effective_source_end,
        "batch_size": batch_size,
        "parallel_mode": "parallel" if parallel else "serial",
        "replica_role": "enabled" if use_replica_role else "disabled",
        "skip_entity_counters": skip_entity_counters,
        "database": os.environ.get("PGDATABASE", "geokrety"),
    }
    run_key = hashlib.sha256(
        stable_json_dumps(payload).encode("utf-8")
    ).hexdigest()
    lock_payload = {
        "requested_start": requested_start,
        "requested_end": requested_end,
        "source_start": effective_source_start,
        "source_end": effective_source_end,
    }
    lock_key = hashlib.sha256(
        stable_json_dumps(lock_payload).encode("utf-8")
    ).hexdigest()
    return {
        **payload,
        "run_key": run_key,
        "lock_key": lock_key,
    }


def build_step_metadata(
    run_identity: dict[str, object],
    phase: str,
    slice_period: Optional[Slice],
) -> dict[str, object]:
    step_payload: dict[str, object] = {
        "phase": phase,
        "step_kind": phase_kind(phase, slice_period),
        "slice_start": (
            format_ts(slice_period.start) if slice_period is not None else None
        ),
        "slice_end": (
            format_ts(slice_period.end) if slice_period is not None else None
        ),
    }
    step_key = hashlib.sha256(
        stable_json_dumps(
            {
                "run_key": run_identity["run_key"],
                **step_payload,
            }
        ).encode("utf-8")
    ).hexdigest()
    return {
        **run_identity,
        **step_payload,
        "step_key": step_key,
    }


def count_step_units(plan: Iterable[tuple[PlanItem, Optional[Slice]]]) -> int:
    total = 0
    for phase_or_group, _slice_period in plan:
        if isinstance(phase_or_group, tuple):
            total += len(phase_or_group)
        else:
            total += 1
    return total


def fetch_completed_step_keys(
    conn: psycopg2.extensions.connection,
    run_key: str,
) -> set[str]:
    with conn.cursor() as cur:
        cur.execute(
            """
            SELECT metadata->>'step_key'
            FROM stats.job_log
            WHERE job_name = %s
              AND status = 'ok'
              AND metadata->>'run_key' = %s
              AND metadata->>'resume_version' = %s
            """,
            (RUNNER_JOB_NAME, run_key, str(RESUME_VERSION)),
        )
        rows = cur.fetchall()
    return {row[0] for row in rows if row[0]}


def clear_resume_markers(
    conn: psycopg2.extensions.connection,
    run_key: str,
) -> int:
    with conn.cursor() as cur:
        cur.execute(
            """
            DELETE FROM stats.job_log
            WHERE job_name = %s
              AND metadata->>'run_key' = %s
            """,
            (RUNNER_JOB_NAME, run_key),
        )
        return cur.rowcount


def insert_resume_marker(
    conn: psycopg2.extensions.connection,
    metadata: dict[str, object],
    started_at: dt.datetime,
    completed_at: dt.datetime,
) -> None:
    with conn.cursor() as cur:
        cur.execute(
            """
            INSERT INTO stats.job_log (
                job_name,
                status,
                metadata,
                started_at,
                completed_at
            )
            VALUES (%s, 'ok', %s::jsonb, %s, %s)
            ON CONFLICT DO NOTHING
            """,
            (
                RUNNER_JOB_NAME,
                stable_json_dumps(metadata),
                started_at,
                completed_at,
            ),
        )


def filter_plan_for_resume(
    plan: Iterable[tuple[PlanItem, Optional[Slice]]],
    run_identity: dict[str, object],
    completed_step_keys: set[str],
) -> tuple[list[tuple[PlanItem, Optional[Slice]]], int]:
    filtered: list[tuple[PlanItem, Optional[Slice]]] = []
    skipped_step_units = 0

    for phase_or_group, slice_period in plan:
        if isinstance(phase_or_group, tuple):
            remaining_phases = []
            for phase in phase_or_group:
                step_key = build_step_metadata(
                    run_identity,
                    phase,
                    slice_period,
                )["step_key"]
                if step_key in completed_step_keys:
                    skipped_step_units += 1
                else:
                    remaining_phases.append(phase)

            if not remaining_phases:
                continue
            if len(remaining_phases) == 1:
                filtered.append((remaining_phases[0], slice_period))
            else:
                filtered.append((tuple(remaining_phases), slice_period))
            continue

        step_key = build_step_metadata(
            run_identity,
            phase_or_group,
            slice_period,
        )["step_key"]
        if step_key in completed_step_keys:
            skipped_step_units += 1
            continue
        filtered.append((phase_or_group, slice_period))

    return filtered, skipped_step_units


def try_acquire_run_lock(
    conn: psycopg2.extensions.connection,
    lock_key: str,
) -> bool:
    key_a, key_b = advisory_lock_keys(lock_key)
    with conn.cursor() as cur:
        cur.execute("SELECT pg_try_advisory_lock(%s, %s)", (key_a, key_b))
        return bool(cur.fetchone()[0])


def release_run_lock(
    conn: psycopg2.extensions.connection,
    lock_key: str,
) -> None:
    key_a, key_b = advisory_lock_keys(lock_key)
    with conn.cursor() as cur:
        cur.execute("SELECT pg_advisory_unlock(%s, %s)", (key_a, key_b))


def parse_bound(
    raw: Optional[str],
    *,
    default: Optional[dt.datetime] = None,
) -> dt.datetime:
    if raw is None:
        if default is None:
            raise ValueError("missing required date bound")
        return default

    value = raw.strip()
    for fmt in ("%Y-%m", "%Y-%m-%d"):
        try:
            parsed = dt.datetime.strptime(value, fmt)
            return parsed.replace(tzinfo=UTC)
        except ValueError:
            continue

    raise ValueError(
        f"unsupported date format: {raw!r}; use YYYY-MM or YYYY-MM-DD"
    )


def default_end() -> dt.datetime:
    today = dt.datetime.now(UTC).date()
    tomorrow = today + dt.timedelta(days=1)
    return dt.datetime.combine(tomorrow, dt.time.min, tzinfo=UTC)


def first_day_next_month(value: dt.datetime) -> dt.datetime:
    year = value.year + (1 if value.month == 12 else 0)
    month = 1 if value.month == 12 else value.month + 1
    return dt.datetime(year, month, 1, tzinfo=UTC)


def iter_month_slices(start: dt.datetime, end: dt.datetime) -> Iterator[Slice]:
    cursor = start
    while cursor < end:
        next_month = first_day_next_month(cursor)
        yield Slice(cursor, min(next_month, end))
        cursor = min(next_month, end)


def connect() -> psycopg2.extensions.connection:
    conn = psycopg2.connect(
        host=os.environ.get("PGHOST"),
        port=os.environ.get("PGPORT", "5432"),
        user=os.environ.get("PGUSER"),
        password=os.environ.get("PGPASSWORD"),
        dbname=os.environ.get("PGDATABASE", "geokrety"),
    )
    conn.autocommit = True
    return conn


def fetch_source_bounds(
    conn: psycopg2.extensions.connection,
) -> tuple[dt.datetime, dt.datetime]:
    with conn.cursor() as cur:
        cur.execute(
            """
            SELECT
              date_trunc('month', min(moved_on_datetime)) AT TIME ZONE 'UTC',
                            date_trunc(
                                'day',
                                max(moved_on_datetime) + interval '1 day'
                            ) AT TIME ZONE 'UTC'
            FROM geokrety.gk_moves
            """
        )
        row = cur.fetchone()

    if row is None or row[0] is None or row[1] is None:
        raise RuntimeError("could not determine gk_moves bounds")

    lower = (
        row[0].replace(tzinfo=UTC)
        if row[0].tzinfo is None
        else row[0].astimezone(UTC)
    )
    upper = (
        row[1].replace(tzinfo=UTC)
        if row[1].tzinfo is None
        else row[1].astimezone(UTC)
    )
    return lower, upper


def fetch_latest_job_log_timing(
    conn: psycopg2.extensions.connection,
    phase: str,
    started_at: dt.datetime,
) -> Optional[int]:
    with conn.cursor() as cur:
        cur.execute(
            """
            SELECT NULLIF(metadata->>'timing_ms', '')::bigint
            FROM stats.job_log
            WHERE job_name = %s
              AND started_at >= %s
            ORDER BY id DESC
            LIMIT 1
            """,
            (phase, started_at),
        )
        row = cur.fetchone()

    if row is None:
        return None
    return row[0]


def execute_phase_sql(
    cur: psycopg2.extensions.cursor,
    phase: str,
    batch_size: int,
    slice_period: Optional[Slice],
) -> dict:
    if slice_period is None:
        cur.execute(
            "SELECT stats.fn_run_snapshot_phase(%s, NULL, %s)::text",
            (phase, batch_size),
        )
    else:
        cur.execute(
            (
                "SELECT stats.fn_run_snapshot_phase("
                "%s, tstzrange(%s, %s, '[)'), %s)::text"
            ),
            (phase, slice_period.start, slice_period.end, batch_size),
        )
    payload = cur.fetchone()[0]
    return json.loads(payload)


def execute_logged_at_author_home_backfill(
    cur: psycopg2.extensions.cursor,
    batch_size: int,
) -> str:
    cur.execute(
        "SELECT stats.fn_backfill_gk_moves_logged_at_author_home(NULL, %s)",
        (batch_size,),
    )
    row = cur.fetchone()
    return "" if row is None or row[0] is None else str(row[0])


def parse_logged_at_author_home_backfill_summary(
    summary: str,
) -> tuple[int, int, int, str]:
    match = re.match(
        r"^Processed (\d+) rows; (\d+) rows updated; (\d+) batch(?:es)? completed; (.+)\.$",
        summary,
    )
    if match is None:
        raise RuntimeError(
            "unexpected logged_at_author_home backfill summary: "
            f"{summary}"
        )
    return (
        int(match.group(1)),
        int(match.group(2)),
        int(match.group(3)),
        match.group(4),
    )


def run_logged_at_author_home_backfill(
    batch_size: int,
    log_file: str,
) -> int:
    event_logger = EventLogger(log_file)
    conn = connect()
    started_at = dt.datetime.now(UTC)
    wall_start = time.monotonic()
    previous_autocommit = conn.autocommit
    lock_key = "backfill_logged_at_author_home_full_history"
    lock_acquired = False
    total_processed = 0
    total_updated = 0
    total_batches = 0
    scope_description = "full-history scope"

    try:
        if previous_autocommit:
            conn.autocommit = False

        lock_acquired = try_acquire_run_lock(conn, lock_key)
        if not lock_acquired:
            raise RuntimeError(
                "another logged_at_author_home backfill is already running"
            )

        target = (
            "stats.fn_backfill_gk_moves_logged_at_author_home"
            f"(NULL, {batch_size})"
        )
        print(f"{EMOJI['info']} Running full-history {target}")
        event_logger.log(
            "logged_at_author_home backfill started: "
            f"full-history, batch_size={batch_size}, replica_role=on, "
            f"lock_key={lock_key}"
        )

        iteration = 0
        while True:
            iteration += 1
            with conn.cursor() as cur:
                cur.execute("SET LOCAL session_replication_role = replica")
                batch_summary = execute_logged_at_author_home_backfill(cur, batch_size)

            processed, updated, batches, scope_description = (
                parse_logged_at_author_home_backfill_summary(batch_summary)
            )
            conn.commit()

            total_processed += processed
            total_updated += updated
            total_batches += batches

            event_logger.log(
                "logged_at_author_home backfill iteration committed: "
                f"iteration={iteration} summary={batch_summary}"
            )

            if processed == 0:
                break

            print(
                f"{EMOJI['running']} Batch {total_batches}: {batch_summary}"
            )

        summary = (
            f"Processed {total_processed} rows; {total_updated} rows updated; "
            f"{total_batches} batches completed; {scope_description}."
        )

        elapsed = time.monotonic() - wall_start
        completed_at = dt.datetime.now(UTC)
        event_logger.log(
            "logged_at_author_home backfill completed: "
            f"summary={summary} wall={elapsed:.2f}s "
            f"started_at={started_at.isoformat()} "
            f"completed_at={completed_at.isoformat()}"
        )
        print(f"{EMOJI['done']} {summary}")
        print(
            f"{EMOJI['stats']} Elapsed: {format_duration(elapsed)}"
        )
        return 0
    except Exception:
        conn.rollback()
        raise
    finally:
        if lock_acquired:
            release_run_lock(conn, lock_key)
            conn.commit()
        if previous_autocommit:
            conn.autocommit = True
        conn.close()
        event_logger.close()


def run_phase(
    conn: psycopg2.extensions.connection,
    phase: str,
    batch_size: int,
    slice_period: Optional[Slice],
    use_replica_role: bool = False,
) -> dict:
    if not use_replica_role:
        with conn.cursor() as cur:
            return execute_phase_sql(cur, phase, batch_size, slice_period)

    previous_autocommit = conn.autocommit
    if previous_autocommit:
        conn.autocommit = False

    try:
        with conn.cursor() as cur:
            cur.execute("SET LOCAL session_replication_role = replica")
            payload = execute_phase_sql(cur, phase, batch_size, slice_period)
        conn.commit()
        return payload
    except Exception:
        conn.rollback()
        raise
    finally:
        if previous_autocommit:
            conn.autocommit = True


def should_use_replica_role(phase: str, enabled: bool) -> bool:
    return enabled and phase in REPLICA_ROLE_PHASES


def progress_worker(state: ProgressState) -> None:
    while not state.stop:
        if state.current_started_at:
            now = time.monotonic()
            current_elapsed = now - state.current_started_at
            overall_elapsed = now - state.overall_started_at

            # Calculate average from completed steps
            average_step = (
                overall_elapsed / state.completed_steps
                if state.completed_steps > 0
                else 0.0
            )
            remaining_steps = state.total_steps - state.completed_steps - 1

            # Task-specific ETA
            task_eta_seconds = current_elapsed + (
                current_elapsed * (remaining_steps / (state.completed_steps + 1))
            ) if state.completed_steps >= 0 else current_elapsed
            now_dt = dt.datetime.now(UTC)
            eta_dt = now_dt + dt.timedelta(seconds=task_eta_seconds)

            phase_for_display = state.current_phase
            # Show active sub-phases while a parallel group is running.
            if state.current_phase.startswith("PARALLEL"):
                with state.parallel_lock:
                    running = list(state.parallel_tasks_started_at.items())
                if running:
                    running.sort(key=lambda item: item[0])
                    chunks = []
                    for phase_name, started_at in running[:3]:
                        elapsed = now - started_at
                        chunks.append(
                            f"{short_phase_name(phase_name)} {format_duration(elapsed)}"
                        )
                    remaining = len(running) - len(chunks)
                    suffix = f" (+{remaining} more)" if remaining > 0 else ""
                    phase_for_display = (
                        f"{state.current_phase} [{'; '.join(chunks)}{suffix}]"
                    )

            if state.dashboard is not None:
                state.dashboard.render(state)
                time.sleep(0.5)
                continue

            # Only update the running task row, not the status line
            state.table_display.print_row(
                state.completed_steps + 1,
                state.total_steps,
                phase_for_display,
                state.current_slice,
                current_elapsed,
                eta_dt,
                is_current=True,
                overwrite=True,
            )
        time.sleep(0.5)


def run_phase_in_thread(
    phase: str,
    batch_size: int,
    slice_period: Optional[Slice],
    use_replica_role: bool,
) -> tuple[str, float, Optional[int], dict, dt.datetime, dt.datetime]:
    """Run a phase with its own DB connection; used for parallel group execution."""
    conn = connect()
    try:
        phase_started_at = dt.datetime.now(UTC)
        wall_start = time.monotonic()
        payload = run_phase(
            conn,
            phase,
            batch_size,
            slice_period,
            use_replica_role=use_replica_role,
        )
        wall_seconds = time.monotonic() - wall_start
        phase_completed_at = dt.datetime.now(UTC)
        timing_ms = fetch_latest_job_log_timing(conn, phase, phase_started_at)
        return (
            phase,
            wall_seconds,
            timing_ms,
            payload,
            phase_started_at,
            phase_completed_at,
        )
    finally:
        conn.close()


# PlanItem is either a single phase name or a tuple of phases to run in parallel.
PlanItem = Union[str, tuple[str, ...]]


def build_plan(
    slices: Iterable[Slice],
    parallel: bool = True,
) -> list[tuple[PlanItem, Optional[Slice]]]:
    plan: list[tuple[PlanItem, Optional[Slice]]] = [
        (phase, None) for phase in PRE_FULL_ONCE_PHASES
    ]
    for slice_period in slices:
        for phase in SEQUENTIAL_MONTHLY_PHASES:
            plan.append((phase, slice_period))
        if parallel:
            plan.append((PARALLEL_MONTHLY_PHASES, slice_period))
        else:
            for phase in PARALLEL_MONTHLY_PHASES:
                plan.append((phase, slice_period))
    plan.extend((phase, None) for phase in POST_FULL_ONCE_PHASES)
    return plan


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Run snapshot backfill phases month by month."
    )
    parser.add_argument(
        "--start",
        help="inclusive start bound in YYYY-MM or YYYY-MM-DD; default 2007-10",
    )
    parser.add_argument(
        "--end",
        help=(
            "exclusive end bound in YYYY-MM or YYYY-MM-DD; "
            "default tomorrow 00:00 UTC"
        ),
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=50000,
        help="batch size passed to fn_run_snapshot_phase",
    )
    parser.add_argument(
        "--backfill-logged-at-author-home",
        action="store_true",
        help=(
            "run the full-history caller-committed "
            "stats.fn_backfill_gk_moves_logged_at_author_home "
            "backfill and exit"
        ),
    )
    parser.add_argument(
        "--skip-entity-counters",
        action="store_true",
        help="skip the full-only fn_snapshot_entity_counters phase",
    )
    parser.add_argument(
        "--no-parallel",
        action="store_true",
        help="run snapshot phases sequentially (disables parallel execution)",
    )
    parser.add_argument(
        "--no-replica-role",
        action="store_true",
        help=(
            "disable transaction-scoped SET LOCAL session_replication_role = "
            "replica for the four full rebuild phases"
        ),
    )
    parser.add_argument(
        "--no-resume",
        action="store_true",
        help="disable reuse of runner-owned step completion markers",
    )
    parser.add_argument(
        "--clear-resume-markers",
        action="store_true",
        help="delete runner-owned completion markers for the exact resolved run key before planning",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="print the execution plan without running it",
    )
    parser.add_argument(
        "--log-file",
        default="docs/database-refactor/run_snapshot_backfill.log",
        help="append real-time phase events to this file",
    )
    parser.add_argument(
        "--dashboard",
        choices=("auto", "rich", "plain"),
        default="auto",
        help="dashboard mode: auto (default), rich, or plain",
    )
    args = parser.parse_args()

    if args.batch_size < 1:
        parser.error("--batch-size must be >= 1")

    if args.backfill_logged_at_author_home:
        return run_logged_at_author_home_backfill(
            args.batch_size,
            args.log_file,
        )

    start = parse_bound(args.start, default=DEFAULT_START)
    end = parse_bound(args.end, default=default_end())
    if end <= start:
        parser.error("--end must be greater than --start")

    parallel = not args.no_parallel
    use_replica_role = not args.no_replica_role
    slices = list(iter_month_slices(start, end))
    plan = build_plan(slices, parallel=parallel)
    if args.skip_entity_counters:
        plan = [
            (phase_or_group, slice_period)
            for phase_or_group, slice_period in plan
            if phase_or_group != "fn_snapshot_entity_counters"
        ]
    original_plan = list(plan)

    if args.dashboard == "rich" and not HAS_RICH:
        parser.error("--dashboard rich requested but python package 'rich' is missing")

    use_rich_dashboard = (
        args.dashboard == "rich"
        or (args.dashboard == "auto" and HAS_RICH and sys.stdout.isatty())
    )

    event_logger = EventLogger(args.log_file)
    conn = connect()
    run_lock_key: Optional[str] = None
    try:
        source_start, source_end = fetch_source_bounds(conn)
        run_identity = build_run_identity(
            start,
            end,
            source_start,
            source_end,
            args.batch_size,
            parallel,
            use_replica_role,
            args.skip_entity_counters,
        )
        if args.clear_resume_markers or not args.dry_run:
            if not try_acquire_run_lock(conn, str(run_identity["lock_key"])):
                raise RuntimeError(
                    "another snapshot backfill is already running for the "
                    "same requested/source bounds"
                )
            run_lock_key = str(run_identity["lock_key"])
        if args.clear_resume_markers:
            deleted_markers = clear_resume_markers(
                conn,
                str(run_identity["run_key"]),
            )
            print(
                f"{EMOJI['info']} Cleared {deleted_markers} resume markers for run_key "
                f"{str(run_identity['run_key'])[:12]}"
            )

        completed_step_keys: set[str] = set()
        skipped_step_units = 0
        if not args.no_resume:
            completed_step_keys = fetch_completed_step_keys(
                conn,
                str(run_identity["run_key"]),
            )
            plan, skipped_step_units = filter_plan_for_resume(
                plan,
                run_identity,
                completed_step_keys,
            )

        total_step_units = count_step_units(original_plan)
        runnable_step_units = count_step_units(plan)

        if args.dry_run:
            print(
                f"{EMOJI['info']} Range: {format_ts(start)} → {format_ts(end)}"
            )
            mode_label = "parallel" if parallel else "serial"
            print(
                f"{EMOJI['info']} Run key: {str(run_identity['run_key'])[:12]} | "
                f"Source: {format_ts(source_start)} → {format_ts(source_end)}"
            )
            print(
                f"{EMOJI['stats']} Plan items: {len(plan)} runnable / {len(original_plan)} total "
                f"({len(slices)} months, {mode_label}, replica-role {'on' if use_replica_role else 'off'}, resume {'off' if args.no_resume else 'on'})"
            )
            print(
                f"{EMOJI['stats']} Step units: {runnable_step_units} runnable / {total_step_units} total; "
                f"skipped via resume: {skipped_step_units}\n"
            )
            for idx, (phase_or_group, slice_period) in enumerate(plan, 1):
                if isinstance(phase_or_group, tuple):
                    phase_type = "PAR"
                    phase_display = (
                        f"PARALLEL ({len(phase_or_group)} phases): "
                        + ", ".join(phase_or_group)
                    )
                elif slice_period is None:
                    phase_type = "FULL"
                    phase_display = phase_or_group
                else:
                    phase_type = "SLICE"
                    phase_display = phase_or_group
                slice_label = (
                    "full"
                    if slice_period is None
                    else slice_period.progress_label
                )
                print(
                    f"  {idx:2}. [{phase_type:5}] {phase_display[:60]:60} "
                    f"{slice_label}"
                )
            return 0

        print(
            f"{EMOJI['info']} Source bounds: {format_ts(source_start)} → {format_ts(source_end)}"
        )
        print(
            f"{EMOJI['info']} Requested run: {format_ts(start)} → {format_ts(end)}"
        )
        print(
            f"{EMOJI['info']} Run key: {str(run_identity['run_key'])[:12]} | Resume: {'off' if args.no_resume else 'on'} | "
            f"Skipped step units: {skipped_step_units}"
        )
        print(
            f"{EMOJI['stats']} Slices: {len(slices)} | Plan items: {len(plan)} | Step units: {runnable_step_units}/{total_step_units} | Batch size: {args.batch_size}\n"
        )

        if not plan:
            print(
                f"{EMOJI['done']} All step units already have completion markers for this exact run. Nothing to execute."
            )
            return 0

        mode_label = "parallel" if parallel else "serial"
        event_logger.log(
            f"run started: {format_ts(start)} -> {format_ts(end)}, "
            f"mode={mode_label}, slices={len(slices)}, steps={len(plan)}, "
            f"batch_size={args.batch_size}, replica_role={'on' if use_replica_role else 'off'}, "
            f"run_key={run_identity['run_key']}, skipped_step_units={skipped_step_units}"
        )

        state = ProgressState(
            total_steps=len(plan),
            overall_started_at=time.monotonic(),
        )
        if use_rich_dashboard:
            state.dashboard = RichDashboard(
                total_steps=len(plan),
                start=start,
                end=end,
                mode_label=mode_label,
                log_file=args.log_file,
            )
            state.dashboard.start_live()

        watcher: Optional[threading.Thread]
        if sys.stdout.isatty():
            watcher = threading.Thread(
                target=progress_worker,
                args=(state,),
                daemon=True,
            )
            watcher.start()
        else:
            watcher = None

        summary: list[
            tuple[str, Optional[Slice], float, Optional[int], dict]
        ] = []

        def emit_event(message: str) -> None:
            event_logger.log(message)
            with state.events_lock:
                state.recent_events.append(message)
                if len(state.recent_events) > 200:
                    state.recent_events = state.recent_events[-200:]
            # In non-TTY mode, print one-line live events to stdout.
            if watcher is None:
                print(message, flush=True)

        for phase_or_group, slice_period in plan:
            slice_label_short = (
                "full" if slice_period is None else slice_period.progress_label
            )
            slice_label_full = (
                "full" if slice_period is None else slice_period.label
            )

            if isinstance(phase_or_group, tuple):
                # ---- PARALLEL GROUP ----
                phases_in_group = phase_or_group
                group_label = f"PARALLEL ({len(phases_in_group)} phases)"
                state.current_phase = group_label
                state.current_slice = slice_label_short
                state.current_started_at = time.monotonic()
                emit_event(f"start {group_label} [{slice_label_short}]")

                with ThreadPoolExecutor(max_workers=len(phases_in_group)) as executor:
                    now_monotonic = time.monotonic()
                    with state.parallel_lock:
                        state.parallel_tasks_started_at = {
                            p: now_monotonic for p in phases_in_group
                        }
                    futures = {
                        executor.submit(
                            run_phase_in_thread,
                            p,
                            args.batch_size,
                            slice_period,
                            should_use_replica_role(p, use_replica_role),
                        ): p
                        for p in phases_in_group
                    }
                    parallel_results: list[
                        tuple[
                            str,
                            float,
                            Optional[int],
                            dict,
                            dt.datetime,
                            dt.datetime,
                        ]
                    ] = []
                    for future in as_completed(futures):
                        phase_name = futures[future]
                        try:
                            result = future.result()
                        except Exception as exc:
                            with state.parallel_lock:
                                state.parallel_tasks_started_at.pop(phase_name, None)
                            emit_event(
                                f"error {phase_name} [{slice_label_short}] {exc}"
                            )
                            raise
                        parallel_results.append(result)
                        phase_done, phase_wall, timing_ms, payload, started_at, completed_at = result
                        with state.parallel_lock:
                            state.parallel_tasks_started_at.pop(phase_done, None)
                        marker_metadata = build_step_metadata(
                            run_identity,
                            phase_done,
                            slice_period,
                        )
                        marker_metadata.update(
                            {
                                "outcome": "completed",
                                "wall_ms": int(phase_wall * 1000),
                                "job_timing_ms": timing_ms,
                                "rows_affected": payload.get("rows_affected"),
                            }
                        )
                        insert_resume_marker(
                            conn,
                            marker_metadata,
                            started_at,
                            completed_at,
                        )
                        timing_display = (
                            "?"
                            if timing_ms is None
                            else f"{timing_ms / 1000.0:.2f}s"
                        )
                        emit_event(
                            f"done {phase_done} [{slice_label_short}] "
                            f"wall={phase_wall:.2f}s job_log={timing_display}"
                        )

                group_wall_seconds = time.monotonic() - state.current_started_at
                emit_event(
                    f"done {group_label} [{slice_label_short}] "
                    f"wall={group_wall_seconds:.2f}s"
                )
                state.completed_steps += 1
                state.steps_timings.append(group_wall_seconds)

                if watcher is not None:
                    if state.dashboard is not None:
                        state.dashboard.render(state)
                    else:
                        now = time.monotonic()
                        overall_elapsed = now - state.overall_started_at
                        average_step = overall_elapsed / state.completed_steps
                        remaining_steps = state.total_steps - state.completed_steps
                        global_eta_seconds = remaining_steps * average_step
                        now_dt = dt.datetime.now(UTC)
                        global_eta_dt = now_dt + dt.timedelta(seconds=global_eta_seconds)
                        eta_dt = now_dt + dt.timedelta(seconds=average_step)

                        state.table_display.finalize_running_line()
                        state.table_display.print_row(
                            state.completed_steps,
                            state.total_steps,
                            group_label,
                            slice_label_full,
                            group_wall_seconds,
                            eta_dt,
                            global_eta_dt=global_eta_dt,
                            is_current=False,
                            overwrite=False,
                        )
                        state.table_display.print_global_status(
                            state.completed_steps,
                            state.total_steps,
                            global_eta_dt,
                            overall_elapsed,
                        )

                # Add each sub-phase to summary in canonical order
                results_by_phase = {r[0]: r for r in parallel_results}
                for p in phases_in_group:
                    p_name, p_wall, p_timing, p_payload, _started_at, _completed_at = results_by_phase[p]
                    summary.append((p_name, slice_period, p_wall, p_timing, p_payload))

            else:
                # ---- SEQUENTIAL STEP ----
                phase = phase_or_group
                state.current_phase = phase
                state.current_slice = slice_label_short
                state.current_started_at = time.monotonic()
                phase_started_at = dt.datetime.now(UTC)
                emit_event(f"start {phase} [{slice_label_short}]")

                wall_started = time.monotonic()
                payload = run_phase(
                    conn,
                    phase,
                    args.batch_size,
                    slice_period,
                    use_replica_role=should_use_replica_role(phase, use_replica_role),
                )
                wall_seconds = time.monotonic() - wall_started
                timing_ms = fetch_latest_job_log_timing(
                    conn,
                    phase,
                    phase_started_at,
                )
                timing_display = (
                    "?"
                    if timing_ms is None
                    else f"{timing_ms / 1000.0:.2f}s"
                )
                emit_event(
                    f"done {phase} [{slice_label_short}] "
                    f"wall={wall_seconds:.2f}s job_log={timing_display}"
                )
                marker_metadata = build_step_metadata(
                    run_identity,
                    phase,
                    slice_period,
                )
                marker_metadata.update(
                    {
                        "outcome": "completed",
                        "wall_ms": int(wall_seconds * 1000),
                        "job_timing_ms": timing_ms,
                        "rows_affected": payload.get("rows_affected"),
                    }
                )
                insert_resume_marker(
                    conn,
                    marker_metadata,
                    phase_started_at,
                    dt.datetime.now(UTC),
                )

                state.completed_steps += 1
                state.steps_timings.append(wall_seconds)

                if watcher is not None:
                    if state.dashboard is not None:
                        state.dashboard.render(state)
                    else:
                        now = time.monotonic()
                        overall_elapsed = now - state.overall_started_at
                        average_step = overall_elapsed / state.completed_steps
                        remaining_steps = state.total_steps - state.completed_steps
                        global_eta_seconds = (
                            overall_elapsed + remaining_steps * average_step
                        )
                        now_dt = dt.datetime.now(UTC)
                        global_eta_dt = now_dt + dt.timedelta(
                            seconds=global_eta_seconds - overall_elapsed
                        )
                        eta_dt = now_dt + dt.timedelta(seconds=average_step)

                        state.table_display.finalize_running_line()
                        state.table_display.print_row(
                            state.completed_steps,
                            state.total_steps,
                            phase,
                            slice_label_full,
                            wall_seconds,
                            eta_dt,
                            global_eta_dt=global_eta_dt,
                            is_current=False,
                            overwrite=False,
                        )
                        state.table_display.print_global_status(
                            state.completed_steps,
                            state.total_steps,
                            global_eta_dt,
                            overall_elapsed,
                        )

                summary.append(
                    (phase, slice_period, wall_seconds, timing_ms, payload)
                )

        state.stop = True
        if watcher is not None:
            watcher.join(timeout=1.0)
        if state.dashboard is not None:
            state.dashboard.render(state)
            state.dashboard.stop_live()

        overall_seconds = time.monotonic() - state.overall_started_at
        event_logger.log(
            f"run completed: duration={overall_seconds:.2f}s, phases={len(summary)}"
        )
        if state.dashboard is None:
            state.table_display.print_summary(overall_seconds, len(summary))
        else:
            throughput = (
                len(summary) / (overall_seconds / 3600.0)
                if overall_seconds > 0
                else 0
            )
            print(
                f"{EMOJI['stats']} Completed {len(summary)} steps in "
                f"{format_duration(overall_seconds)} "
                f"({throughput:.1f} steps/hour)"
            )
        return 0
    finally:
        if run_lock_key is not None:
            release_run_lock(conn, run_lock_key)
        conn.close()
        event_logger.close()


if __name__ == "__main__":
    raise SystemExit(main())
