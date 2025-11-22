# Arquitetura do Genus - Mecanismo de Generics

Este documento explica em detalhes como o Genus usa Go Generics para alcançar type-safety completa.

## Problema que Resolvemos

ORMs tradicionais em Go têm alguns problemas:

### 1. Falta de Type-Safety em Queries

```go
// GORM - não é type-safe
db.Where("name = ?", "Alice").Find(&users)
db.Where("age > ?", "not a number") // Erro só em runtime! 💥
```

### 2. Necessidade de Passar Ponteiros para Slices

```go
// GORM - precisa de *[]T
var users []User
db.Find(&users) // Precisa do & aqui
```

### 3. Uso Excessivo de Reflection

Reflection é lento e pode causar panics em runtime.

## Nossa Solução: Generics em Go 1.18+

### 1. Query Builder Genérico

**Antes (GORM-style):**
```go
func (db *DB) Find(dest interface{}) error {
    // Usa reflection para descobrir o tipo
    // Pode dar panic em runtime
}
```

**Depois (Genus):**
```go
type Builder[T any] struct {
    tableName string
    // ...
}

func (b *Builder[T]) Find(ctx context.Context) ([]T, error) {
    var results []T
    // ... scan rows
    return results, nil // Retorna []T diretamente!
}
```

**Como funciona:**
- O tipo `T` é conhecido em **tempo de compilação**
- O compilador garante que o retorno é sempre `[]T`
- Não precisa de `*[]T` porque criamos o slice internamente

### 2. Função Table Genérica

```go
func Table[T any](g *Genus) *query.Builder[T] {
    var model T
    tableName := getTableName(model)
    return query.NewBuilder[T](
        g.db.Executor(),
        g.db.Dialect(),
        tableName,
    )
}
```

**Fluxo de Type-Safety:**

```go
// 1. Table[User] retorna *Builder[User]
builder := genus.Table[User](db)

// 2. Where retorna *Builder[User] (method chaining)
builder = builder.Where(condition)

// 3. Find retorna []User (type-safe!)
users, err := builder.Find(ctx)
```

O compilador garante que cada passo preserva o tipo `User`.

### 3. Campos Tipados (Type-Safe Fields)

Este é o design mais importante do Genus. Permite queries como:

```go
UserFields.Name.Eq("Alice")  // ✅ String com string
UserFields.Age.Gt(25)        // ✅ Int com int
UserFields.Age.Eq("text")    // ❌ Erro de compilação!
```

**Implementação:**

```go
// 1. Definimos tipos específicos para cada tipo de campo
type StringField struct {
    column string
}

type IntField struct {
    column string
}

// 2. Cada tipo tem seus próprios métodos
func (f StringField) Eq(value string) Condition {
    return Condition{
        Field:    f.column,
        Operator: OpEq,
        Value:    value,  // type-safe: aceita apenas string!
    }
}

func (f IntField) Eq(value int) Condition {
    return Condition{
        Field:    f.column,
        Operator: OpEq,
        Value:    value,  // type-safe: aceita apenas int!
    }
}
```

**Vantagens:**
- Não pode comparar string com int (erro de compilação)
- Autocomplete mostra apenas operadores válidos para o tipo
- Zero reflection - tudo é verificado em compile-time

### 4. Interfaces Genéricas para Comparadores

```go
type Comparador[T any] interface {
    Field
    Eq(value T) Condition
    Ne(value T) Condition
    In(values ...T) Condition
    // ...
}

type ComparadorOrdenavel[T any] interface {
    Comparador[T]
    Gt(value T) Condition
    Gte(value T) Condition
    Lt(value T) Condition
    Lte(value T) Condition
    Between(start, end T) Condition
}
```

**Como isso ajuda:**
- `StringField` implementa `Comparador[string]`
- `IntField` implementa `ComparadorOrdenavel[int]`
- O compilador garante que apenas tipos ordenáveis têm `Gt`, `Lt`, etc.

## Onde Ainda Usamos Reflection

Genus **minimiza** reflection, mas ainda usa em alguns lugares:

### 1. Scanning de Resultados (query/scanner.go)

```go
func scanStruct(rows *sql.Rows, dest interface{}) error {
    // Usa reflection para mapear colunas do DB para campos da struct
    destValue := reflect.ValueOf(dest).Elem()
    // ...
}
```

**Por que aqui?**
- O `database/sql` retorna `*sql.Rows` que não é tipado
- Precisamos mapear dinamicamente as colunas para os campos da struct
- Mas isso é **isolado** - o resto do código é type-safe!

### 2. Operações CRUD (core/db.go)

```go
func (db *DB) Create(ctx context.Context, model interface{}) error {
    // Usa reflection para extrair campos da struct
    columns, values, err := getColumnsAndValues(model)
    // ...
}
```

**Por que aqui?**
- CREATE/UPDATE/DELETE não sabem antecipadamente quais campos a struct tem
- Reflection é necessária para iterar sobre os campos

**Nota importante:** Essas operações CRUD **poderiam** ser feitas com generics também:

```go
func Create[T any](ctx context.Context, model *T) error {
    // ...
}
```

Mas isso exigiria que cada modelo implementasse uma interface específica ou usasse code generation. Por simplicidade, mantemos reflection aqui.

## Comparação: GORM vs Genus

### GORM (Traditional)

```go
type User struct {
    Name string
    Age  int
}

// ❌ Não é type-safe
db.Where("name = ?", "Alice").Find(&users)
db.Where("age > ?", "texto") // Erro só em runtime!

// ❌ Precisa de ponteiro
var users []User
db.Find(&users)

// ❌ Magic strings
db.Where("nonexistent_field = ?", "value") // Erro só em runtime!
```

### Genus (Type-Safe)

```go
type User struct {
    core.Model
    Name string `db:"name"`
    Age  int    `db:"age"`
}

var UserFields = struct {
    Name query.StringField
    Age  query.IntField
}{
    Name: query.NewStringField("name"),
    Age:  query.NewIntField("age"),
}

// ✅ Type-safe
users, err := genus.Table[User](db).
    Where(UserFields.Name.Eq("Alice")).
    Find(ctx)

// ✅ Erro de compilação!
// users, err := genus.Table[User](db).
//     Where(UserFields.Age.Eq("texto")). // Não compila!
//     Find(ctx)

// ✅ Retorna []User diretamente
// users já é []User, não precisa de &
```

## Pattern: Type-Safe Field Definition

Um padrão que recomendamos para definir campos:

```go
// models/user.go
type User struct {
    core.Model
    Name     string `db:"name"`
    Email    string `db:"email"`
    Age      int    `db:"age"`
    IsActive bool   `db:"is_active"`
}

// models/user_fields.go (ou no mesmo arquivo)
var UserFields = struct {
    ID       query.Int64Field
    Name     query.StringField
    Email    query.StringField
    Age      query.IntField
    IsActive query.BoolField
}{
    ID:       query.NewInt64Field("id"),
    Name:     query.NewStringField("name"),
    Email:    query.NewStringField("email"),
    Age:      query.NewIntField("age"),
    IsActive: query.NewBoolField("is_active"),
}
```

**Futuro:** Podemos criar um code generator que gere `UserFields` automaticamente a partir da struct `User`.

## Benefícios do Design

1. **Type-Safety**: Erros pegos em compile-time, não runtime
2. **Performance**: Menos reflection = mais rápido
3. **IDE Support**: Autocomplete funciona perfeitamente
4. **Refactoring**: Renomear campos é seguro (IDE ajuda)
5. **Transparência**: Queries SQL são visíveis e debugáveis

## Limitações e Trade-offs

### 1. Verbosidade dos Field Definitions

**Trade-off:** Precisa definir campos manualmente.

**Solução futura:** Code generation.

### 2. Ainda Usa Reflection em Alguns Lugares

**Trade-off:** Scanning e CRUD usam reflection.

**Alternativa:** Poderíamos exigir que modelos implementem interfaces específicas, mas isso seria mais complexo.

### 3. Requer Go 1.18+

**Trade-off:** Não funciona em versões antigas do Go.

**Justificativa:** Generics são essenciais para o design. Vale a pena exigir Go moderno.

## Evolução Futura

### 1. Code Generation para Fields

```bash
$ genus generate ./models
Generated: models/user_fields.gen.go
```

### 2. CRUD Genérico

```go
func Create[T Model](ctx context.Context, model *T) error {
    // Usa generics + constraints
}
```

### 3. Type-Safe Joins

```go
query.Join[User, Post](
    UserFields.ID,
    PostFields.UserID,
)
```

## Conclusão

Genus usa Go Generics de forma extensiva para:

1. **Eliminar** a necessidade de `*[]T` (retorna `[]T` diretamente)
2. **Garantir** type-safety em queries
3. **Minimizar** reflection (só onde absolutamente necessário)
4. **Melhorar** a experiência do desenvolvedor (autocomplete, refactoring)

O resultado é um ORM que é **seguro**, **rápido** e **fácil de usar**.
