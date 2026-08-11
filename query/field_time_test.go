package query

import (
	"testing"
	"time"
)

var (
	refTime   = time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	laterTime = time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
)

func TestTimeField_ColumnName(t *testing.T) {
	f := NewTimeField("created_at")
	if f.ColumnName() != "created_at" {
		t.Errorf("expected created_at, got %s", f.ColumnName())
	}
}

func TestTimeField_Operators(t *testing.T) {
	f := NewTimeField("created_at")

	tests := []struct {
		name     string
		cond     Condition
		operator Operator
	}{
		{"Eq", f.Eq(refTime), OpEq},
		{"Ne", f.Ne(refTime), OpNe},
		{"Gt", f.Gt(refTime), OpGt},
		{"Gte", f.Gte(refTime), OpGte},
		{"Lt", f.Lt(refTime), OpLt},
		{"Lte", f.Lte(refTime), OpLte},
		{"IsNull", f.IsNull(), OpIsNull},
		{"IsNotNull", f.IsNotNull(), OpIsNotNull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cond.Field != "created_at" {
				t.Errorf("expected field created_at, got %s", tt.cond.Field)
			}
			if tt.cond.Operator != tt.operator {
				t.Errorf("expected operator %s, got %s", tt.operator, tt.cond.Operator)
			}
		})
	}
}

// Os aliases semânticos têm que mapear para os operadores de comparação corretos.
func TestTimeField_SemanticAliases(t *testing.T) {
	f := NewTimeField("created_at")

	tests := []struct {
		name     string
		cond     Condition
		expected Operator
	}{
		{"After", f.After(refTime), OpGt},
		{"Before", f.Before(refTime), OpLt},
		{"OnOrAfter", f.OnOrAfter(refTime), OpGte},
		{"OnOrBefore", f.OnOrBefore(refTime), OpLte},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cond.Operator != tt.expected {
				t.Errorf("%s should map to %s, got %s", tt.name, tt.expected, tt.cond.Operator)
			}
			if tt.cond.Value != refTime {
				t.Errorf("expected value %v, got %v", refTime, tt.cond.Value)
			}
		})
	}
}

func TestTimeField_Between(t *testing.T) {
	cond := NewTimeField("created_at").Between(refTime, laterTime)

	if cond.Operator != OpBetween {
		t.Fatalf("expected BETWEEN, got %s", cond.Operator)
	}

	values, ok := cond.Value.([]time.Time)
	if !ok {
		t.Fatalf("expected []time.Time, got %T", cond.Value)
	}
	if len(values) != 2 || values[0] != refTime || values[1] != laterTime {
		t.Errorf("unexpected range: %v", values)
	}
}

func TestTimeField_InNotIn(t *testing.T) {
	f := NewTimeField("created_at")

	in := f.In(refTime, laterTime)
	if in.Operator != OpIn {
		t.Errorf("expected IN, got %s", in.Operator)
	}
	if values, ok := in.Value.([]time.Time); !ok || len(values) != 2 {
		t.Errorf("expected 2 values, got %v", in.Value)
	}

	notIn := f.NotIn(refTime)
	if notIn.Operator != OpNotIn {
		t.Errorf("expected NOT IN, got %s", notIn.Operator)
	}
}

// Valores fora de UTC têm que ser normalizados. Sem isso, o driver do SQLite
// serializa no fuso local e a comparação textual do SQLite devolve resultado
// errado sem erro nenhum.
func TestTimeField_NormalizesToUTC(t *testing.T) {
	saoPaulo, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Skipf("timezone database unavailable: %v", err)
	}

	local := refTime.In(saoPaulo)
	if local.Location() == time.UTC {
		t.Fatal("test setup: expected a non-UTC time")
	}

	f := NewTimeField("created_at")

	cond := f.Before(local)
	value, ok := cond.Value.(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T", cond.Value)
	}
	if value.Location() != time.UTC {
		t.Errorf("expected UTC, got %v", value.Location())
	}
	if !value.Equal(refTime) {
		t.Errorf("normalization must preserve the instant: %v != %v", value, refTime)
	}

	between := f.Between(local, local.Add(time.Hour))
	values := between.Value.([]time.Time)
	for i, v := range values {
		if v.Location() != time.UTC {
			t.Errorf("Between value %d not normalized: %v", i, v.Location())
		}
	}

	in := f.In(local)
	inValues := in.Value.([]time.Time)
	if inValues[0].Location() != time.UTC {
		t.Errorf("In value not normalized: %v", inValues[0].Location())
	}

	assignment := f.Set(local)
	assigned := assignment.Value.(time.Time)
	if assigned.Location() != time.UTC {
		t.Errorf("Set value not normalized: %v", assigned.Location())
	}
}

func TestOptionalTimeField_NormalizesToUTC(t *testing.T) {
	saoPaulo, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Skipf("timezone database unavailable: %v", err)
	}

	local := refTime.In(saoPaulo)
	f := NewOptionalTimeField("deleted_at")

	value := f.After(local).Value.(time.Time)
	if value.Location() != time.UTC {
		t.Errorf("expected UTC, got %v", value.Location())
	}

	assigned := f.Set(local).Value.(time.Time)
	if assigned.Location() != time.UTC {
		t.Errorf("Set value not normalized: %v", assigned.Location())
	}
}

func TestOptionalTimeField_Operators(t *testing.T) {
	f := NewOptionalTimeField("deleted_at")

	if f.ColumnName() != "deleted_at" {
		t.Errorf("expected deleted_at, got %s", f.ColumnName())
	}

	tests := []struct {
		name     string
		cond     Condition
		operator Operator
	}{
		{"Eq", f.Eq(refTime), OpEq},
		{"Ne", f.Ne(refTime), OpNe},
		{"Gt", f.Gt(refTime), OpGt},
		{"Gte", f.Gte(refTime), OpGte},
		{"Lt", f.Lt(refTime), OpLt},
		{"Lte", f.Lte(refTime), OpLte},
		{"After", f.After(refTime), OpGt},
		{"Before", f.Before(refTime), OpLt},
		{"OnOrAfter", f.OnOrAfter(refTime), OpGte},
		{"OnOrBefore", f.OnOrBefore(refTime), OpLte},
		{"In", f.In(refTime), OpIn},
		{"NotIn", f.NotIn(refTime), OpNotIn},
		{"Between", f.Between(refTime, laterTime), OpBetween},
		{"IsNull", f.IsNull(), OpIsNull},
		{"IsNotNull", f.IsNotNull(), OpIsNotNull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cond.Field != "deleted_at" {
				t.Errorf("expected field deleted_at, got %s", tt.cond.Field)
			}
			if tt.cond.Operator != tt.operator {
				t.Errorf("expected operator %s, got %s", tt.operator, tt.cond.Operator)
			}
		})
	}
}

// BETWEEN com time.Time precisa gerar dois placeholders. interfaceSlice não
// tratava []time.Time, o que fazia a condição ser descartada silenciosamente.
func TestBuilder_TimeBetweenGeneratesSQL(t *testing.T) {
	b := newTestBuilder().Where(NewTimeField("created_at").Between(refTime, laterTime))

	sql, args := b.buildSelectQuery()

	expected := `SELECT * FROM "users" WHERE created_at BETWEEN $1 AND $2`
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[0] != refTime || args[1] != laterTime {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestBuilder_TimeInGeneratesSQL(t *testing.T) {
	b := newTestBuilder().Where(NewTimeField("created_at").In(refTime, laterTime))

	sql, args := b.buildSelectQuery()

	expected := `SELECT * FROM "users" WHERE created_at IN ($1, $2)`
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d: %v", len(args), args)
	}
}

// Regressão: interfaceSlice também não tratava []float64, então Float64Field.In
// e Between geravam SQL com um único placeholder recebendo o slice inteiro.
func TestBuilder_Float64InGeneratesSQL(t *testing.T) {
	b := newTestBuilder().Where(NewFloat64Field("score").In(1.5, 2.5, 3.5))

	sql, args := b.buildSelectQuery()

	expected := `SELECT * FROM "users" WHERE score IN ($1, $2, $3)`
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
	if args[0] != 1.5 || args[2] != 3.5 {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestBuilder_Float64BetweenGeneratesSQL(t *testing.T) {
	b := newTestBuilder().Where(NewFloat64Field("score").Between(1.5, 9.5))

	sql, args := b.buildSelectQuery()

	expected := `SELECT * FROM "users" WHERE score BETWEEN $1 AND $2`
	if sql != expected {
		t.Errorf("expected %q, got %q", expected, sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d: %v", len(args), args)
	}
}
