package query

import "github.com/go-genus/genus/core"

// Scope é uma função que modifica um query builder.
// Scopes podem ser usados para aplicar condições globais a queries.
type Scope[T any] func(*Builder[T]) *Builder[T]

// isSoftDeletable indica se o model T suporta soft delete.
//
// Verifica o value e o ponteiro porque core.SoftDeleteModel implementa a
// interface com receivers de ponteiro: um model que o embute satisfaz
// core.SoftDeletable apenas via *T.
func isSoftDeletable[T any]() bool {
	var zero T
	if _, ok := any(zero).(core.SoftDeletable); ok {
		return true
	}
	_, ok := any(&zero).(core.SoftDeletable)
	return ok
}

// applySoftDeleteScope adiciona automaticamente WHERE deleted_at IS NULL
// para models que implementam SoftDeletable, a menos que scopes estejam desabilitados.
func applySoftDeleteScope[T any](b *Builder[T]) *Builder[T] {
	if !isSoftDeletable[T]() {
		return b
	}

	if b.disableScopes {
		return b
	}

	return b.Where(Condition{
		Field:    deletedAtColumn,
		Operator: OpIsNull,
	})
}
