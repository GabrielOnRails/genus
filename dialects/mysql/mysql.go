package mysql

import (
	"fmt"
)

// Dialect é a implementação do dialeto MySQL.
type Dialect struct{}

// New cria uma nova instância do dialeto MySQL.
func New() *Dialect {
	return &Dialect{}
}

// Placeholder retorna o placeholder para MySQL (?).
// MySQL usa placeholders posicionais simples (?) ao invés de numerados.
func (d *Dialect) Placeholder(n int) string {
	return "?"
}

// QuoteIdentifier adiciona backticks ao identificador.
// MySQL usa backticks (`) para escapar identificadores.
func (d *Dialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf("`%s`", name)
}

// GetType retorna o tipo MySQL para um tipo Go.
func (d *Dialect) GetType(goType string) string {
	typeMap := map[string]string{
		"string":    "VARCHAR(255)",
		"int":       "INT",
		"int64":     "BIGINT",
		"bool":      "BOOLEAN",
		"time.Time": "DATETIME",
		"float64":   "DOUBLE",
		"float32":   "FLOAT",
	}

	if sqlType, ok := typeMap[goType]; ok {
		return sqlType
	}

	return "TEXT" // fallback
}
