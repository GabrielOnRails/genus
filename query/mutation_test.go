package query

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-genus/genus/core"
)

// mockResult implementa sql.Result para os testes de mutação.
type mockResult struct {
	affected int64
	err      error
}

func (r mockResult) LastInsertId() (int64, error) { return 0, nil }
func (r mockResult) RowsAffected() (int64, error) { return r.affected, r.err }

// recordingExecutor captura a última query executada.
type recordingExecutor struct {
	mockExecutor
	query    string
	args     []interface{}
	affected int64
	execErr  error
}

func newRecordingExecutor(affected int64) *recordingExecutor {
	e := &recordingExecutor{affected: affected}
	e.execFn = func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
		e.query = query
		e.args = args
		if e.execErr != nil {
			return nil, e.execErr
		}
		return mockResult{affected: e.affected}, nil
	}
	return e
}

func newRecordingBuilder(affected int64) (*Builder[testUser], *recordingExecutor) {
	exec := newRecordingExecutor(affected)
	return NewBuilder[testUser](exec, &mockDialect{}, newMockLogger(), "users"), exec
}

// ===== Guarda contra mutação sem WHERE =====

func TestUpdate_RefusesWithoutWhere(t *testing.T) {
	b, exec := newRecordingBuilder(0)

	_, err := b.Update(context.Background(), NewStringField("name").Set("x"))
	if err == nil {
		t.Fatal("expected error when updating without WHERE")
	}
	if !strings.Contains(err.Error(), "AllowGlobal") {
		t.Errorf("error should point to AllowGlobal, got: %v", err)
	}
	if exec.query != "" {
		t.Errorf("no query should have been executed, got: %s", exec.query)
	}
}

func TestDelete_RefusesWithoutWhere(t *testing.T) {
	b, exec := newRecordingBuilder(0)

	_, err := b.Delete(context.Background())
	if err == nil {
		t.Fatal("expected error when deleting without WHERE")
	}
	if exec.query != "" {
		t.Errorf("no query should have been executed, got: %s", exec.query)
	}
}

func TestForceDelete_RefusesWithoutWhere(t *testing.T) {
	b, _ := newRecordingBuilder(0)

	if _, err := b.ForceDelete(context.Background()); err == nil {
		t.Fatal("expected error when force deleting without WHERE")
	}
}

func TestUpdate_AllowGlobalPermitsNoWhere(t *testing.T) {
	b, exec := newRecordingBuilder(5)

	affected, err := b.AllowGlobal().Update(context.Background(), NewStringField("name").Set("x"))
	if err != nil {
		t.Fatalf("AllowGlobal should permit update without WHERE: %v", err)
	}
	if affected != 5 {
		t.Errorf("expected 5 rows affected, got %d", affected)
	}
	if strings.Contains(exec.query, "WHERE") {
		t.Errorf("query should have no WHERE, got: %s", exec.query)
	}
}

func TestDelete_AllowGlobalPermitsNoWhere(t *testing.T) {
	b, exec := newRecordingBuilder(3)

	affected, err := b.AllowGlobal().Delete(context.Background())
	if err != nil {
		t.Fatalf("AllowGlobal should permit delete without WHERE: %v", err)
	}
	if affected != 3 {
		t.Errorf("expected 3 rows affected, got %d", affected)
	}
	if exec.query != `DELETE FROM "users"` {
		t.Errorf("unexpected query: %s", exec.query)
	}
}

// AllowGlobal tem que respeitar a imutabilidade do builder.
func TestAllowGlobal_IsImmutable(t *testing.T) {
	b, _ := newRecordingBuilder(0)
	global := b.AllowGlobal()

	if b.allowGlobal {
		t.Error("original builder should not be modified")
	}
	if !global.allowGlobal {
		t.Error("returned builder should have allowGlobal set")
	}
}

// ===== Update =====

func TestUpdate_RequiresAssignments(t *testing.T) {
	b, _ := newRecordingBuilder(0)

	_, err := b.Where(NewIntField("age").Gt(18)).Update(context.Background())
	if err == nil {
		t.Fatal("expected error when no assignment is given")
	}
}

// Em dialetos posicionais, os placeholders do WHERE têm que continuar a
// numeração depois dos do SET.
func TestUpdate_PlaceholderOrdering(t *testing.T) {
	b, exec := newRecordingBuilder(1)

	_, err := b.
		Where(NewStringField("email").Eq("a@test.com")).
		Where(NewIntField("age").Gt(18)).
		Update(context.Background(),
			NewStringField("name").Set("Alice"),
			NewIntField("age").Set(30),
		)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// testUser embute core.Model, então updated_at entra automaticamente como $3.
	expected := `UPDATE "users" SET name = $1, age = $2, updated_at = $3 WHERE email = $4 AND age > $5`
	if exec.query != expected {
		t.Errorf("expected:\n  %s\ngot:\n  %s", expected, exec.query)
	}

	if len(exec.args) != 5 {
		t.Fatalf("expected 5 args, got %d: %v", len(exec.args), exec.args)
	}
	if exec.args[0] != "Alice" || exec.args[1] != 30 {
		t.Errorf("SET args wrong: %v", exec.args[:2])
	}
	if exec.args[3] != "a@test.com" || exec.args[4] != 18 {
		t.Errorf("WHERE args wrong: %v", exec.args[3:])
	}
}

func TestUpdate_AutoSetsUpdatedAt(t *testing.T) {
	b, exec := newRecordingBuilder(1)

	if _, err := b.Where(NewIntField("age").Gt(18)).
		Update(context.Background(), NewStringField("name").Set("x")); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if !strings.Contains(exec.query, "updated_at = $2") {
		t.Errorf("expected automatic updated_at, got: %s", exec.query)
	}
	if _, ok := exec.args[1].(time.Time); !ok {
		t.Errorf("expected time.Time for updated_at, got %T", exec.args[1])
	}
}

// updated_at explícito não deve ser duplicado.
func TestUpdate_ExplicitUpdatedAtNotDuplicated(t *testing.T) {
	b, exec := newRecordingBuilder(1)

	if _, err := b.Where(NewIntField("age").Gt(18)).
		Update(context.Background(), NewTimeField("updated_at").Set(refTime)); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if strings.Count(exec.query, "updated_at") != 1 {
		t.Errorf("updated_at should appear once, got: %s", exec.query)
	}
	if exec.args[0] != refTime {
		t.Errorf("expected explicit time %v, got %v", refTime, exec.args[0])
	}
}

// Model sem coluna updated_at não deve receber a atribuição automática.
func TestUpdate_NoUpdatedAtColumn(t *testing.T) {
	exec := newRecordingExecutor(1)
	b := NewBuilder[testPost](exec, &mockDialect{}, newMockLogger(), "posts")

	if _, err := b.Where(NewInt64Field("user_id").Eq(1)).
		Update(context.Background(), NewStringField("title").Set("novo")); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	expected := `UPDATE "posts" SET title = $1 WHERE user_id = $2`
	if exec.query != expected {
		t.Errorf("expected:\n  %s\ngot:\n  %s", expected, exec.query)
	}
}

func TestUpdate_SetNullOnOptionalField(t *testing.T) {
	exec := newRecordingExecutor(1)
	b := NewBuilder[testPost](exec, &mockDialect{}, newMockLogger(), "posts")

	if _, err := b.Where(NewInt64Field("id").Eq(1)).
		Update(context.Background(), NewOptionalStringField("title").SetNull()); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if exec.args[0] != nil {
		t.Errorf("expected nil value for SetNull, got %v", exec.args[0])
	}
}

func TestUpdate_PropagatesExecError(t *testing.T) {
	b, exec := newRecordingBuilder(0)
	exec.execErr = fmt.Errorf("boom")

	_, err := b.Where(NewIntField("age").Gt(18)).
		Update(context.Background(), NewStringField("name").Set("x"))
	if err == nil {
		t.Fatal("expected error to propagate")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

// Zero linhas afetadas é resultado válido, não erro.
func TestUpdate_ZeroRowsIsNotError(t *testing.T) {
	b, _ := newRecordingBuilder(0)

	affected, err := b.Where(NewIntField("age").Gt(999)).
		Update(context.Background(), NewStringField("name").Set("x"))
	if err != nil {
		t.Fatalf("zero rows should not be an error: %v", err)
	}
	if affected != 0 {
		t.Errorf("expected 0, got %d", affected)
	}
}

// ===== Delete e soft delete =====

func TestDelete_HardDeleteForPlainModel(t *testing.T) {
	b, exec := newRecordingBuilder(2)

	affected, err := b.Where(NewIntField("age").Lt(18)).Delete(context.Background())
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if affected != 2 {
		t.Errorf("expected 2, got %d", affected)
	}

	expected := `DELETE FROM "users" WHERE age < $1`
	if exec.query != expected {
		t.Errorf("expected %q, got %q", expected, exec.query)
	}
}

// Model soft-deletable tem DELETE convertido em UPDATE deleted_at, e o scope
// impede que registros já removidos sejam tocados de novo.
func TestDelete_SoftDeleteBecomesUpdate(t *testing.T) {
	exec := newRecordingExecutor(1)
	b := NewBuilder[softDeleteUser](exec, &mockDialect{}, newMockLogger(), "users")

	affected, err := b.Where(NewStringField("name").Eq("Alice")).Delete(context.Background())
	if err != nil {
		t.Fatalf("soft delete failed: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1, got %d", affected)
	}

	if !strings.HasPrefix(exec.query, `UPDATE "users" SET deleted_at = $1`) {
		t.Errorf("expected UPDATE deleted_at, got: %s", exec.query)
	}
	if !strings.Contains(exec.query, "deleted_at IS NULL") {
		t.Errorf("soft delete should skip already deleted rows, got: %s", exec.query)
	}
	if _, ok := exec.args[0].(time.Time); !ok {
		t.Errorf("expected time.Time for deleted_at, got %T", exec.args[0])
	}
}

// ForceDelete ignora soft delete e alcança registros já removidos.
func TestForceDelete_IgnoresSoftDelete(t *testing.T) {
	exec := newRecordingExecutor(4)
	b := NewBuilder[softDeleteUser](exec, &mockDialect{}, newMockLogger(), "users")

	affected, err := b.Where(NewStringField("name").Eq("Alice")).ForceDelete(context.Background())
	if err != nil {
		t.Fatalf("force delete failed: %v", err)
	}
	if affected != 4 {
		t.Errorf("expected 4, got %d", affected)
	}

	expected := `DELETE FROM "users" WHERE name = $1`
	if exec.query != expected {
		t.Errorf("expected %q, got %q", expected, exec.query)
	}
}

// Update em model soft-deletable não deve atingir registros removidos.
func TestUpdate_SkipsSoftDeletedRows(t *testing.T) {
	exec := newRecordingExecutor(1)
	b := NewBuilder[softDeleteUser](exec, &mockDialect{}, newMockLogger(), "users")

	if _, err := b.Where(NewStringField("name").Eq("Alice")).
		Update(context.Background(), NewStringField("name").Set("Alicia")); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if !strings.Contains(exec.query, "deleted_at IS NULL") {
		t.Errorf("expected soft delete scope, got: %s", exec.query)
	}
}

func TestIsSoftDeletable(t *testing.T) {
	if !isSoftDeletable[softDeleteUser]() {
		t.Error("model embedding core.SoftDeleteModel should be soft deletable")
	}
	if !isSoftDeletable[valueSoftDeletable]() {
		t.Error("value-receiver implementation should be soft deletable")
	}
	if isSoftDeletable[testUser]() {
		t.Error("plain model should not be soft deletable")
	}
}

func TestHasColumn(t *testing.T) {
	if !hasColumn[testUser]("updated_at") {
		t.Error("testUser embeds core.Model, should have updated_at")
	}
	if hasColumn[testPost]("updated_at") {
		t.Error("testPost has no updated_at column")
	}
	// Builder[any] é usado internamente pelos helpers de join.
	if hasColumn[any]("updated_at") {
		t.Error("interface type should report no columns")
	}
}

// ===== Integração com SQLite =====

func TestUpdate_IntegrationAffectsOnlyMatchingRows(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	affected, err := newSQLiteBuilder(db).
		Where(NewIntField("age").Gte(30)).
		Update(ctx, NewStringField("email").Set("updated@test.com"))
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Alice (30) e Charlie (35).
	if affected != 2 {
		t.Fatalf("expected 2 rows affected, got %d", affected)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = 'updated@test.com'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected 2 updated rows in db, got %d", count)
	}

	// Os demais registros seguem intactos.
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email != 'updated@test.com'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("expected 3 untouched rows, got %d", count)
	}
}

func TestDelete_IntegrationRemovesOnlyMatchingRows(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	affected, err := newSQLiteBuilder(db).
		Where(NewIntField("age").Lt(26)).
		Delete(ctx)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Bob (25) e Eve (22).
	if affected != 2 {
		t.Fatalf("expected 2 rows affected, got %d", affected)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("expected 3 remaining rows, got %d", count)
	}
}

// Update com filtro temporal, ponta a ponta contra o banco.
func TestUpdate_IntegrationWithTimeFilter(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// created_at é preenchido pelo default CURRENT_TIMESTAMP, então tudo
	// está antes de agora + 1h e depois de agora - 1h.
	cutoff := time.Now().Add(time.Hour)

	affected, err := newSQLiteBuilder(db).
		Where(NewTimeField("created_at").Before(cutoff)).
		Update(ctx, NewIntField("age").Set(99))
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if affected != 5 {
		t.Fatalf("expected all 5 rows affected, got %d", affected)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE age = 99`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Errorf("expected 5 rows with age 99, got %d", count)
	}
}

func TestUpdate_IntegrationTimeBetween(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)

	affected, err := newSQLiteBuilder(db).
		Where(NewTimeField("created_at").Between(start, end)).
		Update(ctx, NewIntField("age").Set(42))
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if affected != 5 {
		t.Errorf("expected 5 rows in range, got %d", affected)
	}
}

// Soft delete ponta a ponta: a linha continua na tabela, com deleted_at preenchido.
type sqliteSoftUser struct {
	core.SoftDeleteModel
	Name string `db:"name"`
}

func (u sqliteSoftUser) TableName() string { return "soft_users" }

func TestDelete_IntegrationSoftDelete(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(`
		CREATE TABLE soft_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO soft_users (name) VALUES ('Alice'), ('Bob')`); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	b := NewBuilder[sqliteSoftUser](db, sqliteDialectInstance, newMockLogger(), "soft_users")

	affected, err := b.Where(NewStringField("name").Eq("Alice")).Delete(context.Background())
	if err != nil {
		t.Fatalf("soft delete failed: %v", err)
	}
	if affected != 1 {
		t.Fatalf("expected 1 row affected, got %d", affected)
	}

	// A linha continua existindo.
	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM soft_users`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("soft delete should keep the row, got %d rows", total)
	}

	// E está marcada como deletada.
	var deleted int
	if err := db.QueryRow(`SELECT COUNT(*) FROM soft_users WHERE deleted_at IS NOT NULL`).Scan(&deleted); err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 soft deleted row, got %d", deleted)
	}

	// Repetir o delete não afeta nada, porque o scope exclui removidos.
	affected, err = b.Where(NewStringField("name").Eq("Alice")).Delete(context.Background())
	if err != nil {
		t.Fatalf("second soft delete failed: %v", err)
	}
	if affected != 0 {
		t.Errorf("already deleted row should not be affected again, got %d", affected)
	}

	// ForceDelete remove de vez, inclusive o que já estava soft deleted.
	affected, err = b.Where(NewStringField("name").Eq("Alice")).ForceDelete(context.Background())
	if err != nil {
		t.Fatalf("force delete failed: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected force delete to remove 1 row, got %d", affected)
	}

	if err := db.QueryRow(`SELECT COUNT(*) FROM soft_users`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Errorf("expected 1 row left after force delete, got %d", total)
	}
}
