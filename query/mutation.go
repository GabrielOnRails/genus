package query

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-genus/genus/core"
)

const (
	deletedAtColumn = "deleted_at"
	updatedAtColumn = "updated_at"
)

// AllowGlobal libera UPDATE e DELETE sem cláusula WHERE.
//
// Por padrão, Update, Delete e ForceDelete recusam rodar sem WHERE, para que um
// filtro esquecido não afete a tabela inteira. Use este método quando atingir
// todos os registros for realmente a intenção.
//
// IMUTÁVEL: Retorna um novo builder sem modificar o original.
func (b *Builder[T]) AllowGlobal() *Builder[T] {
	newBuilder := b.clone()
	newBuilder.allowGlobal = true
	return newBuilder
}

// Update aplica as atribuições a todos os registros que casam com o WHERE
// e retorna a quantidade de linhas afetadas.
//
// Se o model tem coluna updated_at, ela é atualizada automaticamente, a menos
// que já esteja entre as atribuições passadas. Models soft-deletable não têm
// registros já removidos atualizados.
//
// Zero linhas afetadas não é erro: significa que nenhum registro casou.
//
// Exemplo:
//
//	afetados, err := genus.Table[User](db).
//	    Where(UserFields.LastLogin.Before(corte)).
//	    Update(ctx, UserFields.IsActive.Set(false))
func (b *Builder[T]) Update(ctx context.Context, assignments ...Assignment) (int64, error) {
	if len(assignments) == 0 {
		return 0, fmt.Errorf("update requires at least one assignment")
	}

	if err := b.guardGlobalMutation("UPDATE"); err != nil {
		return 0, err
	}

	// Não atualiza registros já soft-deleted.
	queryBuilder := applySoftDeleteScope(b)

	query, args := queryBuilder.buildUpdateQuery(assignments)
	return queryBuilder.execMutation(ctx, query, args, "update")
}

// Delete remove todos os registros que casam com o WHERE e retorna a
// quantidade de linhas afetadas.
//
// Se o model implementa core.SoftDeletable, executa soft delete
// (UPDATE deleted_at) em vez de DELETE. Use ForceDelete para remover
// definitivamente.
//
// Zero linhas afetadas não é erro: significa que nenhum registro casou.
//
// Exemplo:
//
//	afetados, err := genus.Table[Session](db).
//	    Where(SessionFields.ExpiresAt.Before(time.Now())).
//	    Delete(ctx)
func (b *Builder[T]) Delete(ctx context.Context) (int64, error) {
	if err := b.guardGlobalMutation("DELETE"); err != nil {
		return 0, err
	}

	// Soft delete reaproveita o caminho de Update, que já aplica o scope
	// para não redeletar registros removidos.
	if isSoftDeletable[T]() {
		return b.Update(ctx, Assignment{Field: deletedAtColumn, Value: time.Now().UTC()})
	}

	query, args := b.buildDeleteQuery()
	return b.execMutation(ctx, query, args, "delete")
}

// ForceDelete remove permanentemente os registros que casam com o WHERE,
// ignorando soft delete mesmo se o model implementa core.SoftDeletable.
//
// Alcança também registros já soft-deleted, o que o torna útil para expurgo.
func (b *Builder[T]) ForceDelete(ctx context.Context) (int64, error) {
	if err := b.guardGlobalMutation("DELETE"); err != nil {
		return 0, err
	}

	query, args := b.buildDeleteQuery()
	return b.execMutation(ctx, query, args, "delete")
}

// guardGlobalMutation recusa mutações sem WHERE, a menos que AllowGlobal
// tenha sido chamado explicitamente.
func (b *Builder[T]) guardGlobalMutation(operation string) error {
	if len(b.conditions) > 0 || b.allowGlobal {
		return nil
	}
	return fmt.Errorf(
		"refusing to run %s without a WHERE clause: add Where(...) or call AllowGlobal() to affect every row",
		operation,
	)
}

// buildUpdateQuery constrói a query UPDATE.
func (b *Builder[T]) buildUpdateQuery(assignments []Assignment) (string, []interface{}) {
	var sb strings.Builder

	assignments = b.withUpdatedAt(assignments)
	args := make([]interface{}, 0, len(assignments))

	sb.WriteString("UPDATE ")
	sb.WriteString(b.dialect.QuoteIdentifier(b.tableName))
	sb.WriteString(" SET ")

	setParts := make([]string, len(assignments))
	for i, assignment := range assignments {
		setParts[i] = fmt.Sprintf("%s = %s", assignment.Field, b.dialect.Placeholder(i+1))
		args = append(args, assignment.Value)
	}
	sb.WriteString(strings.Join(setParts, ", "))

	if len(b.conditions) > 0 {
		sb.WriteString(" WHERE ")
		// Placeholders do WHERE continuam a numeração depois dos do SET.
		whereSQL, whereArgs := b.buildWhereClauseFrom(b.conditions, len(assignments)+1)
		sb.WriteString(whereSQL)
		args = append(args, whereArgs...)
	}

	return sb.String(), args
}

// buildDeleteQuery constrói a query DELETE.
func (b *Builder[T]) buildDeleteQuery() (string, []interface{}) {
	var sb strings.Builder
	var args []interface{}

	sb.WriteString("DELETE FROM ")
	sb.WriteString(b.dialect.QuoteIdentifier(b.tableName))

	if len(b.conditions) > 0 {
		sb.WriteString(" WHERE ")
		whereSQL, whereArgs := b.buildWhereClause(b.conditions)
		sb.WriteString(whereSQL)
		args = append(args, whereArgs...)
	}

	return sb.String(), args
}

// withUpdatedAt acrescenta updated_at = agora se o model tiver a coluna
// e o chamador não a tiver atribuído explicitamente.
func (b *Builder[T]) withUpdatedAt(assignments []Assignment) []Assignment {
	if !hasColumn[T](updatedAtColumn) {
		return assignments
	}

	for _, assignment := range assignments {
		if assignment.Field == updatedAtColumn {
			return assignments
		}
	}

	result := make([]Assignment, len(assignments), len(assignments)+1)
	copy(result, assignments)
	return append(result, Assignment{Field: updatedAtColumn, Value: time.Now().UTC()})
}

// execMutation executa a query e retorna as linhas afetadas.
func (b *Builder[T]) execMutation(ctx context.Context, query string, args []interface{}, operation string) (int64, error) {
	start := time.Now()
	result, err := b.executor.ExecContext(ctx, query, args...)
	duration := time.Since(start).Nanoseconds()

	if err != nil {
		b.logger.LogError(query, args, err)
		return 0, fmt.Errorf("failed to %s: %w", operation, err)
	}

	b.logger.LogQuery(query, args, duration)

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// hasColumn indica se o model T tem a coluna informada.
func hasColumn[T any](column string) bool {
	var zero T

	t := reflect.TypeOf(zero)
	if t == nil {
		// T é uma interface (ex: Builder[any]): sem metadados de coluna.
		return false
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}

	for _, col := range core.GetTypeMetadata(t).Columns {
		if col == column {
			return true
		}
	}
	return false
}
