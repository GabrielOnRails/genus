package core

import "time"

// Model é a struct base que deve ser embutida em todos os modelos.
// Usa embedding para fornecer campos comuns.
type Model struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TableNamer é uma interface que os modelos podem implementar
// para especificar o nome da tabela customizado.
type TableNamer interface {
	TableName() string
}

// BeforeCreater é um hook executado antes de criar um registro.
type BeforeCreater interface {
	BeforeCreate() error
}

// AfterFinder é um hook executado após buscar um registro.
type AfterFinder interface {
	AfterFind() error
}
