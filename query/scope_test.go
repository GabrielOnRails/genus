package query

import (
	"testing"
	"time"

	"github.com/go-genus/genus/core"
)

// softDeleteUser implements SoftDeletable
type softDeleteUser struct {
	core.SoftDeleteModel
	Name string `db:"name"`
}

func (u softDeleteUser) TableName() string {
	return "users"
}

// core.SoftDeleteModel implementa SoftDeletable com receivers de ponteiro, então
// um model que o embute satisfaz a interface apenas via *T. Como embutir
// core.SoftDeleteModel é a forma recomendada de habilitar soft delete, o scope
// tem que ser aplicado também nesse caso.
func TestApplySoftDeleteScope_PointerReceiverSoftDeletable(t *testing.T) {
	b := NewBuilder[softDeleteUser](
		&mockExecutor{},
		&mockDialect{},
		newMockLogger(),
		"users",
	)

	result := applySoftDeleteScope(b)

	if len(result.conditions) != 1 {
		t.Fatalf("applySoftDeleteScope should add condition for pointer-receiver SoftDeletable, got %d", len(result.conditions))
	}

	cond, ok := result.conditions[0].(Condition)
	if !ok {
		t.Fatalf("expected Condition, got %T", result.conditions[0])
	}
	if cond.Field != "deleted_at" || cond.Operator != OpIsNull {
		t.Errorf("expected deleted_at IS NULL, got %s %s", cond.Field, cond.Operator)
	}
}

func TestApplySoftDeleteScope_DisabledScopes(t *testing.T) {
	b := NewBuilder[softDeleteUser](
		&mockExecutor{},
		&mockDialect{},
		newMockLogger(),
		"users",
	)
	b.disableScopes = true

	result := applySoftDeleteScope(b)

	if len(result.conditions) != 0 {
		t.Error("applySoftDeleteScope should not add conditions when scopes disabled")
	}
}

func TestApplySoftDeleteScope_NonSoftDeletable(t *testing.T) {
	b := newTestBuilder() // testUser does not implement SoftDeletable
	result := applySoftDeleteScope(b)

	if len(result.conditions) != 0 {
		t.Error("applySoftDeleteScope should not add conditions for non-SoftDeletable types")
	}
}

// Verify softDeleteUser implements SoftDeletable
func TestSoftDeleteUserInterface(t *testing.T) {
	u := &softDeleteUser{}
	var _ core.SoftDeletable = u

	// Test GetDeletedAt
	if u.GetDeletedAt() != nil {
		t.Error("GetDeletedAt should return nil for zero value")
	}

	// Test SetDeletedAt
	now := time.Now()
	u.SetDeletedAt(&now)
	if !u.IsDeleted() {
		t.Error("IsDeleted should return true after SetDeletedAt")
	}

	// Test undelete
	u.SetDeletedAt(nil)
	if u.IsDeleted() {
		t.Error("IsDeleted should return false after SetDeletedAt(nil)")
	}
}
