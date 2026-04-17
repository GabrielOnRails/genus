package tracing

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-genus/genus/core"
)

// MetricsCollector collects database operation metrics.
// It provides counters and histograms compatible with OpenTelemetry conventions.
type MetricsCollector struct {
	// Query metrics
	queryCount  atomic.Int64
	queryErrors atomic.Int64
	execCount   atomic.Int64
	execErrors  atomic.Int64

	// Duration tracking (microseconds)
	mu             sync.Mutex
	queryDurations []int64 // in microseconds
	maxDurations   int     // max stored durations before rotation

	// Operation breakdown
	opCounts sync.Map // map[string]*atomic.Int64

	// Pool metrics callback (optional)
	poolStats func() *PoolMetrics
}

// PoolMetrics contains database connection pool statistics.
type PoolMetrics struct {
	OpenConnections   int
	InUse             int
	Idle              int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxIdleClosed     int64
	MaxLifetimeClosed int64
}

// MetricsSnapshot is a point-in-time snapshot of collected metrics.
type MetricsSnapshot struct {
	// Counters
	TotalQueries int64 `json:"total_queries"`
	TotalExecs   int64 `json:"total_execs"`
	QueryErrors  int64 `json:"query_errors"`
	ExecErrors   int64 `json:"exec_errors"`

	// Latency (microseconds)
	AvgQueryDuration int64 `json:"avg_query_duration_us"`
	P50QueryDuration int64 `json:"p50_query_duration_us"`
	P95QueryDuration int64 `json:"p95_query_duration_us"`
	P99QueryDuration int64 `json:"p99_query_duration_us"`
	MaxQueryDuration int64 `json:"max_query_duration_us"`

	// Per-operation counts
	OperationCounts map[string]int64 `json:"operation_counts"`

	// Pool (nil if no pool stats callback)
	Pool *PoolMetrics `json:"pool,omitempty"`
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		maxDurations: 10000,
	}
}

// SetPoolStatsFunc sets a callback to retrieve connection pool statistics.
// Typically used with sql.DB.Stats().
func (m *MetricsCollector) SetPoolStatsFunc(fn func() *PoolMetrics) {
	m.poolStats = fn
}

// SetPoolStatsFromDB configures pool stats collection from a *sql.DB.
func (m *MetricsCollector) SetPoolStatsFromDB(db *sql.DB) {
	m.poolStats = func() *PoolMetrics {
		stats := db.Stats()
		return &PoolMetrics{
			OpenConnections:   stats.OpenConnections,
			InUse:             stats.InUse,
			Idle:              stats.Idle,
			WaitCount:         stats.WaitCount,
			WaitDuration:      stats.WaitDuration,
			MaxIdleClosed:     stats.MaxIdleClosed,
			MaxLifetimeClosed: stats.MaxLifetimeClosed,
		}
	}
}

// recordQuery records a query execution.
func (m *MetricsCollector) recordQuery(duration time.Duration, err error, operation string) {
	m.queryCount.Add(1)
	if err != nil {
		m.queryErrors.Add(1)
	}

	us := duration.Microseconds()
	m.mu.Lock()
	if len(m.queryDurations) >= m.maxDurations {
		// Rotate: keep the last half
		copy(m.queryDurations, m.queryDurations[m.maxDurations/2:])
		m.queryDurations = m.queryDurations[:m.maxDurations/2]
	}
	m.queryDurations = append(m.queryDurations, us)
	m.mu.Unlock()

	if operation != "" {
		val, _ := m.opCounts.LoadOrStore(operation, &atomic.Int64{})
		val.(*atomic.Int64).Add(1)
	}
}

// recordExec records an exec execution.
func (m *MetricsCollector) recordExec(duration time.Duration, err error, operation string) {
	m.execCount.Add(1)
	if err != nil {
		m.execErrors.Add(1)
	}

	us := duration.Microseconds()
	m.mu.Lock()
	if len(m.queryDurations) >= m.maxDurations {
		copy(m.queryDurations, m.queryDurations[m.maxDurations/2:])
		m.queryDurations = m.queryDurations[:m.maxDurations/2]
	}
	m.queryDurations = append(m.queryDurations, us)
	m.mu.Unlock()

	if operation != "" {
		val, _ := m.opCounts.LoadOrStore(operation, &atomic.Int64{})
		val.(*atomic.Int64).Add(1)
	}
}

// Snapshot returns a point-in-time snapshot of all metrics.
func (m *MetricsCollector) Snapshot() MetricsSnapshot {
	snap := MetricsSnapshot{
		TotalQueries:    m.queryCount.Load(),
		TotalExecs:      m.execCount.Load(),
		QueryErrors:     m.queryErrors.Load(),
		ExecErrors:      m.execErrors.Load(),
		OperationCounts: make(map[string]int64),
	}

	// Calculate percentiles
	m.mu.Lock()
	if len(m.queryDurations) > 0 {
		sorted := make([]int64, len(m.queryDurations))
		copy(sorted, m.queryDurations)
		m.mu.Unlock()

		sortInt64s(sorted)
		n := len(sorted)

		snap.AvgQueryDuration = avgInt64(sorted)
		snap.P50QueryDuration = sorted[n*50/100]
		snap.P95QueryDuration = sorted[n*95/100]
		snap.P99QueryDuration = sorted[min(n*99/100, n-1)]
		snap.MaxQueryDuration = sorted[n-1]
	} else {
		m.mu.Unlock()
	}

	// Collect operation counts
	m.opCounts.Range(func(key, value any) bool {
		snap.OperationCounts[key.(string)] = value.(*atomic.Int64).Load()
		return true
	})

	// Pool stats
	if m.poolStats != nil {
		snap.Pool = m.poolStats()
	}

	return snap
}

// Reset clears all collected metrics.
func (m *MetricsCollector) Reset() {
	m.queryCount.Store(0)
	m.queryErrors.Store(0)
	m.execCount.Store(0)
	m.execErrors.Store(0)

	m.mu.Lock()
	m.queryDurations = m.queryDurations[:0]
	m.mu.Unlock()

	m.opCounts.Range(func(key, _ any) bool {
		m.opCounts.Delete(key)
		return true
	})
}

// MetricsExecutor wraps an Executor and collects metrics on every operation.
type MetricsExecutor struct {
	executor  core.Executor
	collector *MetricsCollector
}

// NewMetricsExecutor creates an executor that records metrics.
func NewMetricsExecutor(executor core.Executor, collector *MetricsCollector) *MetricsExecutor {
	return &MetricsExecutor{
		executor:  executor,
		collector: collector,
	}
}

// ExecContext executes a query and records metrics.
func (me *MetricsExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	op := detectOperation(query)
	start := time.Now()
	result, err := me.executor.ExecContext(ctx, query, args...)
	me.collector.recordExec(time.Since(start), err, op)
	return result, err
}

// QueryContext executes a query and records metrics.
func (me *MetricsExecutor) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	op := detectOperation(query)
	start := time.Now()
	rows, err := me.executor.QueryContext(ctx, query, args...)
	me.collector.recordQuery(time.Since(start), err, op)
	return rows, err
}

// QueryRowContext executes a query row and records metrics.
func (me *MetricsExecutor) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	op := detectOperation(query)
	start := time.Now()
	row := me.executor.QueryRowContext(ctx, query, args...)
	me.collector.recordQuery(time.Since(start), nil, op)
	return row
}

// Collector returns the underlying metrics collector.
func (me *MetricsExecutor) Collector() *MetricsCollector {
	return me.collector
}

// detectOperation extracts the SQL operation type from a query.
func detectOperation(query string) string {
	// Skip leading whitespace
	for i := 0; i < len(query); i++ {
		c := query[i]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			query = query[i:]
			break
		}
	}

	if len(query) < 3 {
		return "UNKNOWN"
	}

	// Check first characters for common operations
	switch query[0] {
	case 'S', 's':
		if len(query) >= 6 && (query[:6] == "SELECT" || query[:6] == "select") {
			return "SELECT"
		}
	case 'I', 'i':
		if len(query) >= 6 && (query[:6] == "INSERT" || query[:6] == "insert") {
			return "INSERT"
		}
	case 'U', 'u':
		if len(query) >= 6 && (query[:6] == "UPDATE" || query[:6] == "update") {
			return "UPDATE"
		}
	case 'D', 'd':
		if len(query) >= 6 && (query[:6] == "DELETE" || query[:6] == "delete") {
			return "DELETE"
		}
	case 'C', 'c':
		if len(query) >= 6 && (query[:6] == "CREATE" || query[:6] == "create") {
			return "CREATE"
		}
	case 'A', 'a':
		if len(query) >= 5 && (query[:5] == "ALTER" || query[:5] == "alter") {
			return "ALTER"
		}
	case 'W', 'w':
		if len(query) >= 4 && (query[:4] == "WITH" || query[:4] == "with") {
			return "SELECT" // CTEs are typically SELECT
		}
	}

	return "OTHER"
}

// sortInt64s sorts a slice of int64 in ascending order (insertion sort for small slices).
func sortInt64s(s []int64) {
	n := len(s)
	for i := 1; i < n; i++ {
		key := s[i]
		j := i - 1
		for j >= 0 && s[j] > key {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = key
	}
}

// avgInt64 computes the average of a non-empty int64 slice.
func avgInt64(s []int64) int64 {
	var sum int64
	for _, v := range s {
		sum += v
	}
	return sum / int64(len(s))
}
