# Genus - Type-Safe ORM para Go

Genus é um ORM (Object-Relational Mapper) de próxima geração para Go que usa **Go Generics** extensivamente para garantir **type-safety** completa em todas as operações de banco de dados.

## Filosofia

- **Mínima Magia**: Praticamente zero reflection em runtime (apenas no scanning de resultados)
- **Type-Safety**: Todas as queries são verificadas em tempo de compilação
- **Transparência**: Queries SQL fáceis de visualizar e debugar
- **Simplicidade**: API fluente e intuitiva
- **Context-Aware**: Todas as funções recebem `context.Context`

## Características Principais

### 1. Retorno Direto de Slices (`[]T`)

Diferente de outros ORMs, Genus retorna `[]T` diretamente, sem precisar de `*[]T`:

```go
// ❌ Outros ORMs
var users []User
db.Find(&users)

// ✅ Genus
users, err := genus.Table[User](db).Find(ctx)
```

### 2. Campos Tipados (Type-Safe Fields)

Defina campos tipados uma vez e use-os de forma type-safe:

```go
var UserFields = struct {
    Name  query.StringField
    Age   query.IntField
    Email query.StringField
}{
    Name:  query.NewStringField("name"),
    Age:   query.NewIntField("age"),
    Email: query.NewStringField("email"),
}

// Uso type-safe - verificado em tempo de compilação!
users, err := genus.Table[User](db).
    Where(UserFields.Name.Eq("Alice")).  // ✅ type-safe
    Where(UserFields.Age.Gt(25)).        // ✅ type-safe
    Find(ctx)
```

### 3. Query Builder Fluente

```go
users, err := genus.Table[User](db).
    Where(UserFields.Age.Gt(18)).
    Where(UserFields.IsActive.Eq(true)).
    OrderByDesc("created_at").
    Limit(10).
    Find(ctx)
```

### 4. Operadores Type-Safe

Cada tipo de campo tem seus próprios operadores:

**StringField:**
- `Eq`, `Ne`, `In`, `NotIn`, `Like`, `NotLike`, `IsNull`, `IsNotNull`

**IntField / Int64Field:**
- `Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`, `Between`, `In`, `NotIn`, `IsNull`, `IsNotNull`

**BoolField:**
- `Eq`, `Ne`, `In`, `NotIn`, `IsNull`, `IsNotNull`

### 5. Queries Complexas (AND/OR)

```go
// AND
users, err := genus.Table[User](db).
    Where(query.And(
        UserFields.Age.Gt(18),
        UserFields.IsActive.Eq(true),
    )).
    Find(ctx)

// OR
users, err := genus.Table[User](db).
    Where(query.Or(
        UserFields.Name.Eq("Alice"),
        UserFields.Age.Gt(30),
    )).
    Find(ctx)
```

## Instalação

```bash
go get github.com/gabrieldias/genus
```

## Quick Start

### 1. Defina seu Modelo

```go
import "github.com/gabrieldias/genus/core"

type User struct {
    core.Model        // Embedded: ID, CreatedAt, UpdatedAt
    Name     string   `db:"name"`
    Email    string   `db:"email"`
    Age      int      `db:"age"`
    IsActive bool     `db:"is_active"`
}
```

### 2. Crie Campos Tipados

```go
import "github.com/gabrieldias/genus/query"

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

### 3. Use!

```go
import "github.com/gabrieldias/genus"

func main() {
    ctx := context.Background()

    // Conecta
    db, err := genus.Open("postgres", "postgresql://...")
    if err != nil {
        log.Fatal(err)
    }

    // Query type-safe
    users, err := genus.Table[User](db).
        Where(UserFields.Name.Eq("Alice")).
        Where(UserFields.Age.Gt(18)).
        Find(ctx)

    // Create
    newUser := &User{Name: "Bob", Email: "bob@example.com", Age: 30}
    err = db.DB().Create(ctx, newUser)

    // Update
    newUser.Age = 31
    err = db.DB().Update(ctx, newUser)

    // Delete
    err = db.DB().Delete(ctx, newUser)
}
```

## Como Funciona o Mecanismo de Generics

### 1. Query Builder Genérico

```go
type Builder[T any] struct {
    executor   core.Executor
    dialect    core.Dialect
    tableName  string
    conditions []interface{}
    // ...
}

func (b *Builder[T]) Find(ctx context.Context) ([]T, error) {
    // Executa query e retorna []T diretamente!
    var results []T
    // ... scan rows into results
    return results, nil
}
```

**Vantagem**: O tipo `T` é conhecido em tempo de compilação, então o compilador garante type-safety.

### 2. Campos Tipados

Cada tipo de campo (`StringField`, `IntField`, etc.) tem métodos que retornam `Condition` tipada:

```go
type StringField struct {
    column string
}

func (f StringField) Eq(value string) Condition {
    return Condition{
        Field:    f.column,
        Operator: OpEq,
        Value:    value,  // type-safe!
    }
}
```

**Vantagem**: O compilador garante que você só pode comparar strings com strings, ints com ints, etc.

### 3. Table Function

```go
func Table[T any](g *Genus) *query.Builder[T] {
    var model T
    tableName := getTableName(model)
    return query.NewBuilder[T](g.db.Executor(), g.db.Dialect(), tableName)
}
```

**Vantagem**: `Table[User](db)` retorna um `*Builder[User]`, garantindo type-safety em toda a cadeia.

## Comparação com Outros ORMs

| Característica | GORM | Ent | **Genus** |
|---------------|------|-----|-----------|
| Type-safe queries | ❌ | ✅ | ✅ |
| Retorna `[]T` | ❌ | ✅ | ✅ |
| Zero codegen | ✅ | ❌ | ✅ |
| Reflection mínimo | ❌ | ✅ | ✅ |
| Campos tipados | ❌ | ✅ | ✅ |
| API fluente | ✅ | ✅ | ✅ |

## Roadmap

- [ ] Suporte para MySQL e SQLite
- [ ] Migrations automáticas
- [ ] Relações (HasMany, BelongsTo, ManyToMany)
- [ ] Code generation para campos tipados
- [ ] Hooks avançados
- [ ] Soft deletes
- [ ] Preloading/Eager loading

## Exemplos

Veja `examples/basic/main.go` para um exemplo completo com todos os recursos.

## Licença

MIT

## Contribuindo

Contribuições são bem-vindas! Por favor, abra uma issue ou PR.
