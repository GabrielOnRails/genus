package query

import (
	"context"
	"database/sql"
	"testing"
)

func TestBuilderEach_SQL(t *testing.T) {
	// Test that Each generates the correct SQL by capturing the query
	var capturedQuery string
	executor := &mockExecutor{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
			capturedQuery = query
			return nil, sql.ErrNoRows
		},
	}

	builder := NewBuilder[testUser](executor, &mockDialect{}, newMockLogger(), "users")

	// Each should call QueryContext — we verify the query was built
	for range builder.Where(StringField{column: "name"}.Eq("Alice")).Each(context.Background()) {
		t.Fatal("should not yield any results on error")
	}

	// The query should have been attempted (even though it returned an error)
	if capturedQuery == "" {
		t.Error("expected query to be executed")
	}
}

func TestBuilderEach2_SQL(t *testing.T) {
	var capturedQuery string
	executor := &mockExecutor{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
			capturedQuery = query
			return nil, sql.ErrNoRows
		},
	}

	builder := NewBuilder[testUser](executor, &mockDialect{}, newMockLogger(), "users")

	var gotErr error
	for _, err := range builder.Each2(context.Background()) {
		gotErr = err
		break
	}

	if capturedQuery == "" {
		t.Error("expected query to be executed")
	}
	if gotErr == nil {
		t.Error("expected error from Each2")
	}
}

func TestFastBuilderEach_SQL(t *testing.T) {
	var capturedQuery string
	executor := &mockExecutor{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
			capturedQuery = query
			return nil, sql.ErrNoRows
		},
	}

	builder := NewFastBuilder[testUser](executor, &mockDialect{}, "users")

	for range builder.Each(context.Background()) {
		t.Fatal("should not yield any results on error")
	}

	if capturedQuery == "" {
		t.Error("expected query to be executed")
	}
}

func TestUltraFastBuilderEach_SQL(t *testing.T) {
	var capturedQuery string
	executor := &mockExecutor{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
			capturedQuery = query
			return nil, sql.ErrNoRows
		},
	}

	builder := NewUltraFastBuilder[testUser](executor, &mockDialect{}, "users")

	for range builder.Each(context.Background()) {
		t.Fatal("should not yield any results on error")
	}

	if capturedQuery == "" {
		t.Error("expected query to be executed")
	}
}

func TestFastBuilderEach2_Error(t *testing.T) {
	executor := &mockExecutor{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
			return nil, sql.ErrConnDone
		},
	}

	builder := NewFastBuilder[testUser](executor, &mockDialect{}, "users")

	var gotErr error
	for _, err := range builder.Each2(context.Background()) {
		gotErr = err
		break
	}

	if gotErr == nil {
		t.Error("expected error from Each2")
	}
}

func TestUltraFastBuilderEach2_Error(t *testing.T) {
	executor := &mockExecutor{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
			return nil, sql.ErrConnDone
		},
	}

	builder := NewUltraFastBuilder[testUser](executor, &mockDialect{}, "users")

	var gotErr error
	for _, err := range builder.Each2(context.Background()) {
		gotErr = err
		break
	}

	if gotErr == nil {
		t.Error("expected error from Each2")
	}
}
