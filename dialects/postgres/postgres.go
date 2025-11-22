package postgres

import (
	"fmt"
)

// Dialect é a implementação do dialeto PostgreSQL.
type Dialect struct{}

// New cria uma nova instância do dialeto PostgreSQL.
func New() *Dialect {
	return &Dialect{}
}

// Placeholder retorna o placeholder para PostgreSQL ($1, $2, etc).
func (d *Dialect) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

// QuoteIdentifier adiciona aspas duplas ao identificador.
func (d *Dialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

// GetType retorna o tipo PostgreSQL para um tipo Go.
func (d *Dialect) GetType(goType string) string {
	typeMap := map[string]string{
		"string":    "VARCHAR(255)",
		"int":       "INTEGER",
		"int64":     "BIGINT",
		"bool":      "BOOLEAN",
		"time.Time": "TIMESTAMP",
		"float64":   "DOUBLE PRECISION",
		"float32":   "REAL",
	}

	if sqlType, ok := typeMap[goType]; ok {
		return sqlType
	}

	return "TEXT" // fallback
}
