package core

import (
	"context"
	"database/sql"
)

// Dialect define a interface para diferentes dialetos de banco de dados.
// Cada dialeto (PostgreSQL, MySQL, SQLite) implementa esta interface.
type Dialect interface {
	// Placeholder retorna o placeholder para a posição n (ex: $1 para PostgreSQL, ? para MySQL)
	Placeholder(n int) string

	// QuoteIdentifier adiciona aspas ao identificador (ex: "users" para PostgreSQL)
	QuoteIdentifier(name string) string

	// GetType retorna o tipo SQL para um tipo Go
	GetType(goType string) string
}

// Executor é a interface que pode executar queries.
// Implementada por *sql.DB e *sql.Tx.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Scanner é a interface para tipos que podem escanear valores do banco.
type Scanner interface {
	Scan(dest ...interface{}) error
}
