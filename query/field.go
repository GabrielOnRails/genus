package query

// Field é a interface base para todos os tipos de campos.
// Cada campo conhece seu nome de coluna no banco de dados.
type Field interface {
	ColumnName() string
}

// Comparador genérico para criar condições.
type Comparador[T any] interface {
	Field
	Eq(value T) Condition
	Ne(value T) Condition
	In(values ...T) Condition
	NotIn(values ...T) Condition
	IsNull() Condition
	IsNotNull() Condition
}

// ComparadorOrdenavel adiciona operadores de comparação.
type ComparadorOrdenavel[T any] interface {
	Comparador[T]
	Gt(value T) Condition
	Gte(value T) Condition
	Lt(value T) Condition
	Lte(value T) Condition
	Between(start, end T) Condition
}

// StringField representa um campo string com operadores específicos.
type StringField struct {
	column string
}

func NewStringField(column string) StringField {
	return StringField{column: column}
}

func (f StringField) ColumnName() string {
	return f.column
}

func (f StringField) Eq(value string) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpEq,
		Value:    value,
	}
}

func (f StringField) Ne(value string) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNe,
		Value:    value,
	}
}

func (f StringField) In(values ...string) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIn,
		Value:    values,
	}
}

func (f StringField) NotIn(values ...string) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNotIn,
		Value:    values,
	}
}

func (f StringField) IsNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNull,
	}
}

func (f StringField) IsNotNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNotNull,
	}
}

func (f StringField) Like(pattern string) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpLike,
		Value:    pattern,
	}
}

func (f StringField) NotLike(pattern string) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNotLike,
		Value:    pattern,
	}
}

// IntField representa um campo int com operadores numéricos.
type IntField struct {
	column string
}

func NewIntField(column string) IntField {
	return IntField{column: column}
}

func (f IntField) ColumnName() string {
	return f.column
}

func (f IntField) Eq(value int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpEq,
		Value:    value,
	}
}

func (f IntField) Ne(value int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNe,
		Value:    value,
	}
}

func (f IntField) Gt(value int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpGt,
		Value:    value,
	}
}

func (f IntField) Gte(value int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpGte,
		Value:    value,
	}
}

func (f IntField) Lt(value int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpLt,
		Value:    value,
	}
}

func (f IntField) Lte(value int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpLte,
		Value:    value,
	}
}

func (f IntField) In(values ...int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIn,
		Value:    values,
	}
}

func (f IntField) NotIn(values ...int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNotIn,
		Value:    values,
	}
}

func (f IntField) Between(start, end int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpBetween,
		Value:    []int{start, end},
	}
}

func (f IntField) IsNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNull,
	}
}

func (f IntField) IsNotNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNotNull,
	}
}

// Int64Field representa um campo int64.
type Int64Field struct {
	column string
}

func NewInt64Field(column string) Int64Field {
	return Int64Field{column: column}
}

func (f Int64Field) ColumnName() string {
	return f.column
}

func (f Int64Field) Eq(value int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpEq,
		Value:    value,
	}
}

func (f Int64Field) Ne(value int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNe,
		Value:    value,
	}
}

func (f Int64Field) Gt(value int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpGt,
		Value:    value,
	}
}

func (f Int64Field) Gte(value int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpGte,
		Value:    value,
	}
}

func (f Int64Field) Lt(value int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpLt,
		Value:    value,
	}
}

func (f Int64Field) Lte(value int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpLte,
		Value:    value,
	}
}

func (f Int64Field) In(values ...int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIn,
		Value:    values,
	}
}

func (f Int64Field) NotIn(values ...int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNotIn,
		Value:    values,
	}
}

func (f Int64Field) Between(start, end int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpBetween,
		Value:    []int64{start, end},
	}
}

func (f Int64Field) IsNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNull,
	}
}

func (f Int64Field) IsNotNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNotNull,
	}
}

// BoolField representa um campo booleano.
type BoolField struct {
	column string
}

func NewBoolField(column string) BoolField {
	return BoolField{column: column}
}

func (f BoolField) ColumnName() string {
	return f.column
}

func (f BoolField) Eq(value bool) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpEq,
		Value:    value,
	}
}

func (f BoolField) Ne(value bool) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNe,
		Value:    value,
	}
}

func (f BoolField) IsNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNull,
	}
}

func (f BoolField) IsNotNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNotNull,
	}
}

func (f BoolField) In(values ...bool) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIn,
		Value:    values,
	}
}

func (f BoolField) NotIn(values ...bool) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNotIn,
		Value:    values,
	}
}

// --- Campos Opcionais para suporte a NULL ---

// OptionalStringField representa um campo string que pode ser NULL.
// Usa core.Optional[string] para tipagem segura.
type OptionalStringField struct {
	column string
}

func NewOptionalStringField(column string) OptionalStringField {
	return OptionalStringField{column: column}
}

func (f OptionalStringField) ColumnName() string {
	return f.column
}

func (f OptionalStringField) Eq(value string) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpEq,
		Value:    value,
	}
}

func (f OptionalStringField) Ne(value string) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNe,
		Value:    value,
	}
}

func (f OptionalStringField) In(values ...string) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIn,
		Value:    values,
	}
}

func (f OptionalStringField) NotIn(values ...string) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNotIn,
		Value:    values,
	}
}

func (f OptionalStringField) IsNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNull,
	}
}

func (f OptionalStringField) IsNotNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNotNull,
	}
}

func (f OptionalStringField) Like(pattern string) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpLike,
		Value:    pattern,
	}
}

func (f OptionalStringField) NotLike(pattern string) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNotLike,
		Value:    pattern,
	}
}

// OptionalIntField representa um campo int que pode ser NULL.
type OptionalIntField struct {
	column string
}

func NewOptionalIntField(column string) OptionalIntField {
	return OptionalIntField{column: column}
}

func (f OptionalIntField) ColumnName() string {
	return f.column
}

func (f OptionalIntField) Eq(value int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpEq,
		Value:    value,
	}
}

func (f OptionalIntField) Ne(value int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNe,
		Value:    value,
	}
}

func (f OptionalIntField) Gt(value int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpGt,
		Value:    value,
	}
}

func (f OptionalIntField) Gte(value int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpGte,
		Value:    value,
	}
}

func (f OptionalIntField) Lt(value int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpLt,
		Value:    value,
	}
}

func (f OptionalIntField) Lte(value int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpLte,
		Value:    value,
	}
}

func (f OptionalIntField) In(values ...int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIn,
		Value:    values,
	}
}

func (f OptionalIntField) NotIn(values ...int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNotIn,
		Value:    values,
	}
}

func (f OptionalIntField) Between(start, end int) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpBetween,
		Value:    []int{start, end},
	}
}

func (f OptionalIntField) IsNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNull,
	}
}

func (f OptionalIntField) IsNotNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNotNull,
	}
}

// OptionalInt64Field representa um campo int64 que pode ser NULL.
type OptionalInt64Field struct {
	column string
}

func NewOptionalInt64Field(column string) OptionalInt64Field {
	return OptionalInt64Field{column: column}
}

func (f OptionalInt64Field) ColumnName() string {
	return f.column
}

func (f OptionalInt64Field) Eq(value int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpEq,
		Value:    value,
	}
}

func (f OptionalInt64Field) Ne(value int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNe,
		Value:    value,
	}
}

func (f OptionalInt64Field) Gt(value int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpGt,
		Value:    value,
	}
}

func (f OptionalInt64Field) Gte(value int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpGte,
		Value:    value,
	}
}

func (f OptionalInt64Field) Lt(value int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpLt,
		Value:    value,
	}
}

func (f OptionalInt64Field) Lte(value int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpLte,
		Value:    value,
	}
}

func (f OptionalInt64Field) In(values ...int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIn,
		Value:    values,
	}
}

func (f OptionalInt64Field) NotIn(values ...int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNotIn,
		Value:    values,
	}
}

func (f OptionalInt64Field) Between(start, end int64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpBetween,
		Value:    []int64{start, end},
	}
}

func (f OptionalInt64Field) IsNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNull,
	}
}

func (f OptionalInt64Field) IsNotNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNotNull,
	}
}

// OptionalBoolField representa um campo booleano que pode ser NULL.
type OptionalBoolField struct {
	column string
}

func NewOptionalBoolField(column string) OptionalBoolField {
	return OptionalBoolField{column: column}
}

func (f OptionalBoolField) ColumnName() string {
	return f.column
}

func (f OptionalBoolField) Eq(value bool) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpEq,
		Value:    value,
	}
}

func (f OptionalBoolField) Ne(value bool) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNe,
		Value:    value,
	}
}

func (f OptionalBoolField) IsNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNull,
	}
}

func (f OptionalBoolField) IsNotNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNotNull,
	}
}

func (f OptionalBoolField) In(values ...bool) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIn,
		Value:    values,
	}
}

func (f OptionalBoolField) NotIn(values ...bool) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNotIn,
		Value:    values,
	}
}

// OptionalFloat64Field representa um campo float64 que pode ser NULL.
type OptionalFloat64Field struct {
	column string
}

func NewOptionalFloat64Field(column string) OptionalFloat64Field {
	return OptionalFloat64Field{column: column}
}

func (f OptionalFloat64Field) ColumnName() string {
	return f.column
}

func (f OptionalFloat64Field) Eq(value float64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpEq,
		Value:    value,
	}
}

func (f OptionalFloat64Field) Ne(value float64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNe,
		Value:    value,
	}
}

func (f OptionalFloat64Field) Gt(value float64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpGt,
		Value:    value,
	}
}

func (f OptionalFloat64Field) Gte(value float64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpGte,
		Value:    value,
	}
}

func (f OptionalFloat64Field) Lt(value float64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpLt,
		Value:    value,
	}
}

func (f OptionalFloat64Field) Lte(value float64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpLte,
		Value:    value,
	}
}

func (f OptionalFloat64Field) In(values ...float64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIn,
		Value:    values,
	}
}

func (f OptionalFloat64Field) NotIn(values ...float64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpNotIn,
		Value:    values,
	}
}

func (f OptionalFloat64Field) Between(start, end float64) Condition {
	return Condition{
		Field:    f.column,
		Operator: OpBetween,
		Value:    []float64{start, end},
	}
}

func (f OptionalFloat64Field) IsNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNull,
	}
}

func (f OptionalFloat64Field) IsNotNull() Condition {
	return Condition{
		Field:    f.column,
		Operator: OpIsNotNull,
	}
}
