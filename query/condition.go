package query

// Operator representa um operador SQL.
type Operator string

const (
	OpEq        Operator = "="
	OpNe        Operator = "!="
	OpGt        Operator = ">"
	OpGte       Operator = ">="
	OpLt        Operator = "<"
	OpLte       Operator = "<="
	OpLike      Operator = "LIKE"
	OpNotLike   Operator = "NOT LIKE"
	OpIn        Operator = "IN"
	OpNotIn     Operator = "NOT IN"
	OpBetween   Operator = "BETWEEN"
	OpIsNull    Operator = "IS NULL"
	OpIsNotNull Operator = "IS NOT NULL"
)

// Condition representa uma condição WHERE.
type Condition struct {
	Field    string
	Operator Operator
	Value    interface{}
}

// LogicalOperator representa operadores lógicos (AND, OR).
type LogicalOperator string

const (
	LogicalAnd LogicalOperator = "AND"
	LogicalOr  LogicalOperator = "OR"
)

// ConditionGroup agrupa múltiplas condições com um operador lógico.
type ConditionGroup struct {
	Conditions []interface{} // pode ser Condition ou ConditionGroup
	Operator   LogicalOperator
}

// And combina condições com AND.
func And(conditions ...Condition) ConditionGroup {
	conds := make([]interface{}, len(conditions))
	for i, c := range conditions {
		conds[i] = c
	}
	return ConditionGroup{
		Conditions: conds,
		Operator:   LogicalAnd,
	}
}

// Or combina condições com OR.
func Or(conditions ...Condition) ConditionGroup {
	conds := make([]interface{}, len(conditions))
	for i, c := range conditions {
		conds[i] = c
	}
	return ConditionGroup{
		Conditions: conds,
		Operator:   LogicalOr,
	}
}

// Not cria uma condição NOT (usando != para simplificar).
func Not(condition Condition) Condition {
	switch condition.Operator {
	case OpEq:
		condition.Operator = OpNe
	case OpNe:
		condition.Operator = OpEq
	case OpIsNull:
		condition.Operator = OpIsNotNull
	case OpIsNotNull:
		condition.Operator = OpIsNull
	case OpIn:
		condition.Operator = OpNotIn
	case OpNotIn:
		condition.Operator = OpIn
	}
	return condition
}
