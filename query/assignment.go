package query

import "time"

// Assignment representa uma atribuição da cláusula SET de um UPDATE.
//
// Assignments são criados a partir dos fields tipados, o que garante que
// tanto o nome da coluna quanto o tipo do valor sejam validados em tempo
// de compilação:
//
//	genus.Table[User](db).
//	    Where(UserFields.LastLogin.Before(corte)).
//	    Update(ctx, UserFields.IsActive.Set(false))
type Assignment struct {
	Field string
	Value interface{}
}

// Set atribui um valor à coluna.
func (f StringField) Set(value string) Assignment {
	return Assignment{Field: f.column, Value: value}
}

// Set atribui um valor à coluna.
func (f IntField) Set(value int) Assignment {
	return Assignment{Field: f.column, Value: value}
}

// Set atribui um valor à coluna.
func (f Int64Field) Set(value int64) Assignment {
	return Assignment{Field: f.column, Value: value}
}

// Set atribui um valor à coluna.
func (f BoolField) Set(value bool) Assignment {
	return Assignment{Field: f.column, Value: value}
}

// Set atribui um valor à coluna.
func (f Float64Field) Set(value float64) Assignment {
	return Assignment{Field: f.column, Value: value}
}

// Set atribui um valor à coluna, normalizado para UTC.
func (f TimeField) Set(value time.Time) Assignment {
	return Assignment{Field: f.column, Value: value.UTC()}
}

// Set atribui um valor à coluna.
func (f OptionalStringField) Set(value string) Assignment {
	return Assignment{Field: f.column, Value: value}
}

// SetNull atribui NULL à coluna.
func (f OptionalStringField) SetNull() Assignment {
	return Assignment{Field: f.column, Value: nil}
}

// Set atribui um valor à coluna.
func (f OptionalIntField) Set(value int) Assignment {
	return Assignment{Field: f.column, Value: value}
}

// SetNull atribui NULL à coluna.
func (f OptionalIntField) SetNull() Assignment {
	return Assignment{Field: f.column, Value: nil}
}

// Set atribui um valor à coluna.
func (f OptionalInt64Field) Set(value int64) Assignment {
	return Assignment{Field: f.column, Value: value}
}

// SetNull atribui NULL à coluna.
func (f OptionalInt64Field) SetNull() Assignment {
	return Assignment{Field: f.column, Value: nil}
}

// Set atribui um valor à coluna.
func (f OptionalBoolField) Set(value bool) Assignment {
	return Assignment{Field: f.column, Value: value}
}

// SetNull atribui NULL à coluna.
func (f OptionalBoolField) SetNull() Assignment {
	return Assignment{Field: f.column, Value: nil}
}

// Set atribui um valor à coluna.
func (f OptionalFloat64Field) Set(value float64) Assignment {
	return Assignment{Field: f.column, Value: value}
}

// SetNull atribui NULL à coluna.
func (f OptionalFloat64Field) SetNull() Assignment {
	return Assignment{Field: f.column, Value: nil}
}

// Set atribui um valor à coluna, normalizado para UTC.
func (f OptionalTimeField) Set(value time.Time) Assignment {
	return Assignment{Field: f.column, Value: value.UTC()}
}

// SetNull atribui NULL à coluna.
func (f OptionalTimeField) SetNull() Assignment {
	return Assignment{Field: f.column, Value: nil}
}
