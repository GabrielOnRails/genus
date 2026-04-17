package query

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/go-genus/genus/core"
)

// Each returns an iterator that yields results one at a time.
// Rows are scanned lazily — only one row is held in memory at a time.
// The underlying database rows are closed automatically when iteration completes
// or when the loop breaks early.
//
// Usage:
//
//	for user := range genus.Table[User](db).Where(...).Each(ctx) {
//	    fmt.Println(user.Name)
//	}
//
// Note: If a scan error occurs, iteration stops silently.
// Use Each2 for explicit error handling.
func (b *Builder[T]) Each(ctx context.Context) iter.Seq[T] {
	return func(yield func(T) bool) {
		queryBuilder := applySoftDeleteScope(b)
		query, args := queryBuilder.buildSelectQuery()

		start := time.Now()
		rows, err := queryBuilder.executor.QueryContext(ctx, query, args...)
		duration := time.Since(start).Nanoseconds()

		if err != nil {
			b.logger.LogError(query, args, err)
			return
		}
		defer rows.Close()

		b.logger.LogQuery(query, args, duration)

		for rows.Next() {
			var item T
			if err := scanStruct(rows, &item); err != nil {
				return
			}

			// Hook AfterFind
			if af, ok := any(&item).(core.AfterFinder); ok {
				if err := af.AfterFind(); err != nil {
					return
				}
			}

			if !yield(item) {
				return
			}
		}
	}
}

// Each2 returns an iterator that yields results and errors one at a time.
// This is the error-aware version of Each.
//
// Usage:
//
//	for user, err := range genus.Table[User](db).Where(...).Each2(ctx) {
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Println(user.Name)
//	}
func (b *Builder[T]) Each2(ctx context.Context) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		queryBuilder := applySoftDeleteScope(b)
		query, args := queryBuilder.buildSelectQuery()

		start := time.Now()
		rows, err := queryBuilder.executor.QueryContext(ctx, query, args...)
		duration := time.Since(start).Nanoseconds()

		if err != nil {
			b.logger.LogError(query, args, err)
			var zero T
			yield(zero, fmt.Errorf("failed to execute query: %w", err))
			return
		}
		defer rows.Close()

		b.logger.LogQuery(query, args, duration)

		for rows.Next() {
			var item T
			if err := scanStruct(rows, &item); err != nil {
				var zero T
				if !yield(zero, fmt.Errorf("failed to scan row: %w", err)) {
					return
				}
				continue
			}

			// Hook AfterFind
			if af, ok := any(&item).(core.AfterFinder); ok {
				if err := af.AfterFind(); err != nil {
					var zero T
					if !yield(zero, fmt.Errorf("AfterFind hook failed: %w", err)) {
						return
					}
					continue
				}
			}

			if !yield(item, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			var zero T
			yield(zero, fmt.Errorf("rows iteration error: %w", err))
		}
	}
}

// Each returns an iterator that yields results one at a time using the fast scanner.
// See Builder.Each for usage details.
func (b *FastBuilder[T]) Each(ctx context.Context) iter.Seq[T] {
	return func(yield func(T) bool) {
		query := b.buildSQL()

		rows, err := b.executor.QueryContext(ctx, query, b.args...)
		if err != nil {
			return
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return
		}

		numCols := len(columns)
		scanValues := make([]interface{}, numCols)
		var placeholder interface{}

		for rows.Next() {
			item, err := b.scanOneFast(rows, columns, numCols, scanValues, &placeholder)
			if err != nil {
				return
			}
			if !yield(item) {
				return
			}
		}
	}
}

// Each2 returns an iterator that yields results and errors one at a time using the fast scanner.
// See Builder.Each2 for usage details.
func (b *FastBuilder[T]) Each2(ctx context.Context) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		query := b.buildSQL()

		rows, err := b.executor.QueryContext(ctx, query, b.args...)
		if err != nil {
			var zero T
			yield(zero, fmt.Errorf("query failed: %w", err))
			return
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			var zero T
			yield(zero, fmt.Errorf("failed to get columns: %w", err))
			return
		}

		numCols := len(columns)
		scanValues := make([]interface{}, numCols)
		var placeholder interface{}

		for rows.Next() {
			item, err := b.scanOneFast(rows, columns, numCols, scanValues, &placeholder)
			if err != nil {
				var zero T
				if !yield(zero, err) {
					return
				}
				continue
			}
			if !yield(item, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			var zero T
			yield(zero, fmt.Errorf("rows iteration error: %w", err))
		}
	}
}

// Each returns an iterator that yields results one at a time.
// Uses the registered scan function if available for zero-reflection performance.
// See Builder.Each for usage details.
func (b *UltraFastBuilder[T]) Each(ctx context.Context) iter.Seq[T] {
	return func(yield func(T) bool) {
		query := b.buildSQL()

		rows, err := b.executor.QueryContext(ctx, query, b.args...)
		if err != nil {
			return
		}
		defer rows.Close()

		if b.scanFunc != nil {
			for rows.Next() {
				item, err := b.scanFunc(rows)
				if err != nil {
					return
				}
				if !yield(item) {
					return
				}
			}
			return
		}

		// Fallback to reflection
		for rows.Next() {
			item, err := b.scanOneReflection(rows)
			if err != nil {
				return
			}
			if !yield(item) {
				return
			}
		}
	}
}

// Each2 returns an iterator that yields results and errors one at a time.
// Uses the registered scan function if available for zero-reflection performance.
// See Builder.Each2 for usage details.
func (b *UltraFastBuilder[T]) Each2(ctx context.Context) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		query := b.buildSQL()

		rows, err := b.executor.QueryContext(ctx, query, b.args...)
		if err != nil {
			var zero T
			yield(zero, fmt.Errorf("query failed: %w", err))
			return
		}
		defer rows.Close()

		if b.scanFunc != nil {
			for rows.Next() {
				item, err := b.scanFunc(rows)
				if err != nil {
					var zero T
					if !yield(zero, err) {
						return
					}
					continue
				}
				if !yield(item, nil) {
					return
				}
			}
		} else {
			for rows.Next() {
				item, err := b.scanOneReflection(rows)
				if err != nil {
					var zero T
					if !yield(zero, err) {
						return
					}
					continue
				}
				if !yield(item, nil) {
					return
				}
			}
		}

		if err := rows.Err(); err != nil {
			var zero T
			yield(zero, fmt.Errorf("rows iteration error: %w", err))
		}
	}
}
