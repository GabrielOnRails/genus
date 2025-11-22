package sqlite

import (
	"fmt"
)

// Dialect é a implementação do dialeto SQLite.
type Dialect struct{}

// New cria uma nova instância do dialeto SQLite.
func New() *Dialect {
	return &Dialect{}
}

// Placeholder retorna o placeholder para SQLite (?).
// SQLite usa placeholders posicionais simples (?) ao invés de numerados.
func (d *Dialect) Placeholder(n int) string {
	return "?"
}

// QuoteIdentifier adiciona aspas duplas ao identificador.
// SQLite usa aspas duplas (") para escapar identificadores.
func (d *Dialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

// GetType retorna o tipo SQLite para um tipo Go.
// SQLite tem um sistema de tipos mais flexível que outros bancos.
func (d *Dialect) GetType(goType string) string {
	typeMap := map[string]string{
		"string":    "TEXT",
		"int":       "INTEGER",
		"int64":     "INTEGER",
		"bool":      "INTEGER", // SQLite não tem tipo boolean nativo
		"time.Time": "DATETIME",
		"float64":   "REAL",
		"float32":   "REAL",
	}

	if sqlType, ok := typeMap[goType]; ok {
		return sqlType
	}

	return "TEXT" // fallback
}
