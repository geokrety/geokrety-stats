package replay

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ProgressTracker tracks and displays real-time statistics during replay.
type ProgressTracker struct {
	startTime       time.Time
	processed       atomic.Int64
	errors          atomic.Int64
	skipped         atomic.Int64 // halted moves (non-scoreable, anonymous, etc.)
	lastLogTime     atomic.Int64
	batchStartTime  time.Time
	batchProcessed  int64
	batchErrors     int64

	mu                     sync.Mutex
	refreshInterval        time.Duration
	totalMoves             int64 // optional: if known beforehand
	distinctGKIDs          map[int64]bool
	totalPointsAwarded     int64
	currentMoveDatetime    time.Time
	logTypeCounts          map[int]int64 // counts by pipeline.LogType
	lastSnapshotAt         time.Time
}

// NewProgressTracker creates a new progress tracker.
func NewProgressTracker(refreshInterval time.Duration) *ProgressTracker {
	if refreshInterval <= 0 {
		refreshInterval = 500 * time.Millisecond
	}
	return &ProgressTracker{
		startTime:       time.Now(),
		refreshInterval: refreshInterval,
		batchStartTime:  time.Now(),
		distinctGKIDs:   make(map[int64]bool),
		logTypeCounts:   make(map[int]int64),
		lastSnapshotAt:  time.Now(),
	}
}

// RecordMove increments the processed counter.
func (pt *ProgressTracker) RecordMove(success bool) {
	if success {
		pt.processed.Add(1)
	} else {
		pt.errors.Add(1)
	}
}

// RecordMoveResult records a move's result for detailed tracking.
func (pt *ProgressTracker) RecordMoveResult(moveID int64, gkID int64, awards int, loggedAt time.Time, logType int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.processed.Add(1)  // Count this as a successful processed move
	pt.distinctGKIDs[gkID] = true
	pt.totalPointsAwarded += int64(awards)
	pt.currentMoveDatetime = loggedAt
	pt.logTypeCounts[logType]++ // Track by log type
}

// RecordError increments the error counter.
func (pt *ProgressTracker) RecordError() {
	pt.errors.Add(1)
}

// RecordSkipped increments the skipped counter and updates log type counts.
func (pt *ProgressTracker) RecordSkipped(logType int) {
	pt.mu.Lock()
	pt.skipped.Add(1)
	pt.logTypeCounts[logType]++
	pt.mu.Unlock()
}

// RecordMoveLogType increments the log type counter for successfully processed moves.
func (pt *ProgressTracker) RecordMoveLogType(logType int) {
	pt.mu.Lock()
	pt.logTypeCounts[logType]++
	pt.mu.Unlock()
}

// RecordBatch updates batch statistics.
func (pt *ProgressTracker) RecordBatch(processed, errors int64) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.batchProcessed = processed
	pt.batchErrors = errors
	pt.batchStartTime = time.Now()
}

// SetTotal sets the total expected moves (optional, for progress percentage).
func (pt *ProgressTracker) SetTotal(total int64) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.totalMoves = total
}

// Start begins the live progress display in a goroutine.
// It returns a function to call that will stop the display and return statistics.
func (pt *ProgressTracker) Start(ctx context.Context) func() *Statistics {
	ticker := time.NewTicker(pt.refreshInterval)
	done := make(chan struct{})
	stop := make(chan struct{})
	var once sync.Once

	go func() {
		defer ticker.Stop()
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case <-stop:
				return
			case <-ticker.C:
				pt.displayProgress()
			}
		}
	}()

	return func() *Statistics {
		once.Do(func() { close(stop) })
		<-done
		return pt.GetStatistics()
	}
}

// displayProgress renders the current progress to stdout.
func (pt *ProgressTracker) displayProgress() {
	processed := pt.processed.Load()
	errors := pt.errors.Load()
	elapsed := time.Since(pt.startTime)

	pt.mu.Lock()
	total := pt.totalMoves
	distinctGKCount := int64(len(pt.distinctGKIDs))
	totalPoints := pt.totalPointsAwarded
	currentMoveDate := pt.currentMoveDatetime
	pt.mu.Unlock()

	rate := float64(processed) / elapsed.Seconds()
	if elapsed.Seconds() < 0.001 {
		rate = 0
	}

	// Calculate ETA on first 10% of moves (if total is known)
	var eta string
	if total > 0 && processed > 0 {
		remaining := total - processed
		secRemaining := float64(remaining) / rate
		eta = fmt.Sprintf(" ETA: %s", formatDuration(time.Duration(int64(secRemaining)) * time.Second))
	}

	// Build progress bar if total is known
	var progress string
	if total > 0 {
		pct := int(processed * 100 / total)
		progress = fmt.Sprintf(" [%d%%]", pct)
	}

	// Format current move date if available
	var dateStr string
	if !currentMoveDate.IsZero() {
		dateStr = fmt.Sprintf(" | Date: %s", currentMoveDate.Format("2006-01-02"))
	}

	skipped := pt.skipped.Load()

	// Clear line and print status with additional metrics
	fmt.Printf("\r\033[K[%s] Processed: %d | Skipped: %d | Errors: %d | Rate: %.1f/sec | GKs: %d | Points: %d%s%s%s\033[0m",
		formatDuration(elapsed),
		processed,
		skipped,
		errors,
		rate,
		distinctGKCount,
		totalPoints,
		dateStr,
		progress,
		eta,
	)

	pt.mu.Lock()
	if time.Since(pt.lastSnapshotAt) >= 30*time.Second {
		fmt.Print("\n")
		pt.lastSnapshotAt = time.Now()
	}
	pt.mu.Unlock()
}

// GetStatistics returns the current replay statistics.
func (pt *ProgressTracker) GetStatistics() *Statistics {
	elapsed := time.Since(pt.startTime)
	processed := pt.processed.Load()
	errors := pt.errors.Load()
	skipped := pt.skipped.Load()

	var rate float64
	if elapsed.Seconds() > 0 {
		rate = float64(processed) / elapsed.Seconds()
	}

	pt.mu.Lock()
	distinctGKCount := int64(len(pt.distinctGKIDs))
	totalPoints := pt.totalPointsAwarded
	logTypeCounts := make(map[int]int64)
	for k, v := range pt.logTypeCounts {
		logTypeCounts[k] = v
	}
	pt.mu.Unlock()

	grandTotal := processed + errors + skipped
	var successRate float64
	if grandTotal > 0 {
		successRate = float64(processed) / float64(grandTotal)
	}

	return &Statistics{
		StartTime:        pt.startTime,
		EndTime:          time.Now(),
		Elapsed:          elapsed,
		Processed:        processed,
		Errors:           errors,
		Skipped:          skipped,
		Rate:             rate,
		Total:            grandTotal,
		SuccessRate:      successRate,
		DistinctGKs:      distinctGKCount,
		TotalPointsAwarded: totalPoints,
		LogTypeCounts:    logTypeCounts,
	}
}

// Statistics holds the final replay statistics.
type Statistics struct {
	StartTime       time.Time
	EndTime         time.Time
	Elapsed         time.Duration
	Processed       int64
	Errors          int64
	Skipped         int64
	Rate            float64
	Total           int64
	SuccessRate     float64
	DistinctGKs     int64
	TotalPointsAwarded int64
	LogTypeCounts   map[int]int64 // move counts by log type
}

// String formats statistics for display.
func (s *Statistics) String() string {
	skippedPct := 0.0
	if s.Total > 0 {
		skippedPct = float64(s.Skipped) / float64(s.Total) * 100
	}

	// Helper to build a properly padded line
	buildLine := func(content string) string {
		if len(content) > 56 {
			content = content[:56]
		}
		padding := strings.Repeat(" ", 56-len(content))
		return fmt.Sprintf("║ %s%s ║", content, padding)
	}

	lines := []string{
		"╔══════════════════════════════════════════════════════════╗",
		"║               REPLAY SESSION SUMMARY                     ║",
		"╠══════════════════════════════════════════════════════════╣",
	}

	// Add data rows
	lines = append(lines, buildLine(fmt.Sprintf("Duration:         %s", formatDuration(s.Elapsed))))
	lines = append(lines, buildLine(fmt.Sprintf("Total Moves:      %d", s.Total)))
	lines = append(lines, buildLine(fmt.Sprintf("Successful:       %d (%.1f%%)", s.Processed, s.SuccessRate*100)))
	lines = append(lines, buildLine(fmt.Sprintf("Errors:           %d (%.1f%%)", s.Errors, (float64(s.Errors)/float64(s.Total))*100)))
	lines = append(lines, buildLine(fmt.Sprintf("Skipped:          %d (%.1f%%)", s.Skipped, skippedPct)))
	lines = append(lines, buildLine(fmt.Sprintf("Rate:             %.2f moves/sec", s.Rate)))
	lines = append(lines, buildLine(fmt.Sprintf("Distinct GKs:     %d", s.DistinctGKs)))
	lines = append(lines, buildLine(fmt.Sprintf("Total Points:     %d", s.TotalPointsAwarded)))
	lines = append(lines, buildLine(fmt.Sprintf("Start:            %s", s.StartTime.Format("2006-01-02 15:04:05"))))
	lines = append(lines, buildLine(fmt.Sprintf("End:              %s", s.EndTime.Format("2006-01-02 15:04:05"))))

	// Add log type breakdown if present
	if len(s.LogTypeCounts) > 0 {
		lines = append(lines, buildLine("Breakdown by Log Type:"))
		for logType, count := range s.LogTypeCounts {
			var typeName string
			switch logType {
			case 0:
				typeName = "Drop"
			case 1:
				typeName = "Grab"
			case 2:
				typeName = "Comment"
			case 3:
				typeName = "Seen"
			case 4:
				typeName = "Archived"
			case 5:
				typeName = "Dip"
			default:
				typeName = fmt.Sprintf("Type%d", logType)
			}
			lines = append(lines, buildLine(fmt.Sprintf("   %s: %d", typeName, count)))
		}
	}

	lines = append(lines, "╚══════════════════════════════════════════════════════════╝")

	return strings.Join(lines, "\n")
}

// formatDuration formats a duration for display (e.g., "1h 23m 45s").
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
