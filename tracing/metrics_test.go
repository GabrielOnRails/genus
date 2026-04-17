package tracing

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

// --- Mock executor for metrics tests ---

type metricsTestExecutor struct {
	delay time.Duration
	err   error
}

func (e *metricsTestExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if e.delay > 0 {
		time.Sleep(e.delay)
	}
	if e.err != nil {
		return nil, e.err
	}
	return &metricsTestResult{}, nil
}

func (e *metricsTestExecutor) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if e.delay > 0 {
		time.Sleep(e.delay)
	}
	if e.err != nil {
		return nil, e.err
	}
	return nil, fmt.Errorf("no rows in mock")
}

func (e *metricsTestExecutor) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}

type metricsTestResult struct{}

func (r *metricsTestResult) LastInsertId() (int64, error) { return 0, nil }
func (r *metricsTestResult) RowsAffected() (int64, error) { return 1, nil }

// --- Tests ---

func TestMetricsCollector_RecordQuery(t *testing.T) {
	mc := NewMetricsCollector()
	mc.recordQuery(100*time.Microsecond, nil, "SELECT")
	mc.recordQuery(200*time.Microsecond, nil, "SELECT")
	mc.recordQuery(50*time.Microsecond, fmt.Errorf("timeout"), "SELECT")

	snap := mc.Snapshot()

	if snap.TotalQueries != 3 {
		t.Errorf("expected 3 queries, got %d", snap.TotalQueries)
	}

	if snap.QueryErrors != 1 {
		t.Errorf("expected 1 query error, got %d", snap.QueryErrors)
	}

	if snap.OperationCounts["SELECT"] != 3 {
		t.Errorf("expected 3 SELECT operations, got %d", snap.OperationCounts["SELECT"])
	}
}

func TestMetricsCollector_RecordExec(t *testing.T) {
	mc := NewMetricsCollector()
	mc.recordExec(100*time.Microsecond, nil, "INSERT")
	mc.recordExec(200*time.Microsecond, nil, "UPDATE")
	mc.recordExec(50*time.Microsecond, fmt.Errorf("constraint"), "INSERT")

	snap := mc.Snapshot()

	if snap.TotalExecs != 3 {
		t.Errorf("expected 3 execs, got %d", snap.TotalExecs)
	}

	if snap.ExecErrors != 1 {
		t.Errorf("expected 1 exec error, got %d", snap.ExecErrors)
	}

	if snap.OperationCounts["INSERT"] != 2 {
		t.Errorf("expected 2 INSERT operations, got %d", snap.OperationCounts["INSERT"])
	}

	if snap.OperationCounts["UPDATE"] != 1 {
		t.Errorf("expected 1 UPDATE operation, got %d", snap.OperationCounts["UPDATE"])
	}
}

func TestMetricsCollector_Percentiles(t *testing.T) {
	mc := NewMetricsCollector()

	// Record 100 queries with increasing durations
	for i := 1; i <= 100; i++ {
		mc.recordQuery(time.Duration(i)*time.Microsecond, nil, "SELECT")
	}

	snap := mc.Snapshot()

	if snap.P50QueryDuration == 0 {
		t.Error("P50 should not be zero")
	}

	if snap.P95QueryDuration == 0 {
		t.Error("P95 should not be zero")
	}

	if snap.P99QueryDuration == 0 {
		t.Error("P99 should not be zero")
	}

	if snap.MaxQueryDuration != 100 {
		t.Errorf("expected max duration 100us, got %d", snap.MaxQueryDuration)
	}

	// P95 should be >= P50
	if snap.P95QueryDuration < snap.P50QueryDuration {
		t.Errorf("P95 (%d) should be >= P50 (%d)", snap.P95QueryDuration, snap.P50QueryDuration)
	}
}

func TestMetricsCollector_Reset(t *testing.T) {
	mc := NewMetricsCollector()
	mc.recordQuery(100*time.Microsecond, nil, "SELECT")
	mc.recordExec(100*time.Microsecond, nil, "INSERT")

	mc.Reset()
	snap := mc.Snapshot()

	if snap.TotalQueries != 0 {
		t.Errorf("expected 0 queries after reset, got %d", snap.TotalQueries)
	}
	if snap.TotalExecs != 0 {
		t.Errorf("expected 0 execs after reset, got %d", snap.TotalExecs)
	}
	if len(snap.OperationCounts) != 0 {
		t.Errorf("expected empty operation counts after reset, got %d entries", len(snap.OperationCounts))
	}
}

func TestMetricsCollector_PoolStats(t *testing.T) {
	mc := NewMetricsCollector()
	mc.SetPoolStatsFunc(func() *PoolMetrics {
		return &PoolMetrics{
			OpenConnections: 10,
			InUse:           5,
			Idle:            5,
		}
	})

	snap := mc.Snapshot()
	if snap.Pool == nil {
		t.Fatal("expected pool stats")
	}
	if snap.Pool.OpenConnections != 10 {
		t.Errorf("expected 10 open connections, got %d", snap.Pool.OpenConnections)
	}
}

func TestMetricsExecutor_ExecContext(t *testing.T) {
	mc := NewMetricsCollector()
	executor := &metricsTestExecutor{}
	me := NewMetricsExecutor(executor, mc)

	ctx := context.Background()
	_, err := me.ExecContext(ctx, "INSERT INTO users (name) VALUES ($1)", "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snap := mc.Snapshot()
	if snap.TotalExecs != 1 {
		t.Errorf("expected 1 exec, got %d", snap.TotalExecs)
	}
	if snap.OperationCounts["INSERT"] != 1 {
		t.Errorf("expected 1 INSERT, got %d", snap.OperationCounts["INSERT"])
	}
}

func TestMetricsExecutor_QueryContextError(t *testing.T) {
	mc := NewMetricsCollector()
	executor := &metricsTestExecutor{err: fmt.Errorf("connection refused")}
	me := NewMetricsExecutor(executor, mc)

	ctx := context.Background()
	_, err := me.QueryContext(ctx, "SELECT * FROM users")
	if err == nil {
		t.Fatal("expected error")
	}

	snap := mc.Snapshot()
	if snap.TotalQueries != 1 {
		t.Errorf("expected 1 query, got %d", snap.TotalQueries)
	}
	if snap.QueryErrors != 1 {
		t.Errorf("expected 1 query error, got %d", snap.QueryErrors)
	}
}

func TestDetectOperation(t *testing.T) {
	tests := []struct {
		query    string
		expected string
	}{
		{"SELECT * FROM users", "SELECT"},
		{"select * from users", "SELECT"},
		{"INSERT INTO users VALUES (1)", "INSERT"},
		{"UPDATE users SET name = 'foo'", "UPDATE"},
		{"DELETE FROM users WHERE id = 1", "DELETE"},
		{"CREATE TABLE users (id INT)", "CREATE"},
		{"ALTER TABLE users ADD COLUMN age INT", "ALTER"},
		{"WITH cte AS (SELECT 1) SELECT * FROM cte", "SELECT"},
		{"  SELECT * FROM users", "SELECT"},
		{"TRUNCATE users", "OTHER"},
		{"", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := detectOperation(tt.query)
			if got != tt.expected {
				t.Errorf("detectOperation(%q) = %q, want %q", tt.query, got, tt.expected)
			}
		})
	}
}

func TestMetricsCollector_DurationRotation(t *testing.T) {
	mc := NewMetricsCollector()
	mc.maxDurations = 100

	// Record more than maxDurations
	for i := 0; i < 150; i++ {
		mc.recordQuery(time.Duration(i)*time.Microsecond, nil, "SELECT")
	}

	mc.mu.Lock()
	count := len(mc.queryDurations)
	mc.mu.Unlock()

	// Should have rotated, so count should be less than 150
	if count >= 150 {
		t.Errorf("expected rotation to reduce stored durations, got %d", count)
	}
}
