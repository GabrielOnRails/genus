# Genus - Guia de Uso

## Índice

1. [Instalação](#instalação)
2. [Configuração](#configuração)
3. [Definindo Modelos](#definindo-modelos)
4. [Criando Campos Tipados](#criando-campos-tipados)
5. [Operações CRUD](#operações-crud)
6. [Queries Type-Safe](#queries-type-safe)
7. [Transações](#transações)
8. [Exemplos Avançados](#exemplos-avançados)

## Instalação

```bash
go get github.com/gabrieldias/genus
```

## Configuração

### Conectando ao Banco de Dados

```go
import (
    "github.com/gabrieldias/genus"
    _ "github.com/lib/pq" // Driver PostgreSQL
)

func main() {
    db, err := genus.Open(
        "postgres",
        "host=localhost user=myuser password=mypass dbname=mydb sslmode=disable",
    )
    if err != nil {
        log.Fatal(err)
    }
    defer db.DB().Executor().(*sql.DB).Close()
}
```

## Definindo Modelos

### Modelo Básico

```go
import "github.com/gabrieldias/genus/core"

type User struct {
    core.Model              // Fornece: ID, CreatedAt, UpdatedAt
    Name       string       `db:"name"`
    Email      string       `db:"email"`
    Age        int          `db:"age"`
    IsActive   bool         `db:"is_active"`
}
```

### Modelo com Nome de Tabela Customizado

```go
type User struct {
    core.Model
    Name  string `db:"name"`
    Email string `db:"email"`
}

func (u User) TableName() string {
    return "app_users" // Ao invés de "user" (padrão)
}
```

### Modelo com Hooks

```go
type User struct {
    core.Model
    Name  string `db:"name"`
    Email string `db:"email"`
}

// Hook executado antes de criar
func (u *User) BeforeCreate() error {
    // Validações, defaults, etc
    if u.Email == "" {
        return fmt.Errorf("email is required")
    }
    return nil
}

// Hook executado após buscar
func (u *User) AfterFind() error {
    // Processamento pós-fetch
    u.Email = strings.ToLower(u.Email)
    return nil
}
```

## Criando Campos Tipados

### Definição Básica

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

### Tipos de Campos Disponíveis

| Tipo Go | Field Type | Operadores |
|---------|-----------|------------|
| `string` | `StringField` | Eq, Ne, Like, NotLike, In, NotIn, IsNull, IsNotNull |
| `int` | `IntField` | Eq, Ne, Gt, Gte, Lt, Lte, Between, In, NotIn, IsNull, IsNotNull |
| `int64` | `Int64Field` | Eq, Ne, Gt, Gte, Lt, Lte, Between, In, NotIn, IsNull, IsNotNull |
| `bool` | `BoolField` | Eq, Ne, In, NotIn, IsNull, IsNotNull |

## Operações CRUD

### Create

```go
user := &User{
    Name:     "Alice",
    Email:    "alice@example.com",
    Age:      28,
    IsActive: true,
}

err := db.DB().Create(ctx, user)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created user with ID: %d\n", user.ID)
```

### Update

```go
user.Age = 29
err := db.DB().Update(ctx, user)
if err != nil {
    log.Fatal(err)
}
```

### Delete

```go
err := db.DB().Delete(ctx, user)
if err != nil {
    log.Fatal(err)
}
```

## Queries Type-Safe

### Find All

```go
users, err := genus.Table[User](db).Find(ctx)
if err != nil {
    log.Fatal(err)
}

for _, user := range users {
    fmt.Printf("%s (%s)\n", user.Name, user.Email)
}
```

### Where (Condição Simples)

```go
users, err := genus.Table[User](db).
    Where(UserFields.Name.Eq("Alice")).
    Find(ctx)
```

### Where (Múltiplas Condições - AND)

```go
users, err := genus.Table[User](db).
    Where(UserFields.Age.Gt(18)).
    Where(UserFields.IsActive.Eq(true)).
    Find(ctx)

// Equivale a: WHERE age > 18 AND is_active = true
```

### Where (Condições Complexas - AND/OR)

```go
// AND explícito
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

// Combinação
users, err := genus.Table[User](db).
    Where(query.And(
        UserFields.IsActive.Eq(true),
        query.Or(
            UserFields.Name.Eq("Alice"),
            UserFields.Age.Gt(30),
        ),
    )).
    Find(ctx)
```

### Operadores

#### String

```go
// Igual
UserFields.Name.Eq("Alice")

// Diferente
UserFields.Name.Ne("Bob")

// LIKE
UserFields.Email.Like("%@example.com")

// NOT LIKE
UserFields.Email.NotLike("%spam%")

// IN
UserFields.Name.In("Alice", "Bob", "Charlie")

// NOT IN
UserFields.Name.NotIn("Spam", "Fake")

// IS NULL
UserFields.Name.IsNull()

// IS NOT NULL
UserFields.Name.IsNotNull()
```

#### Int/Int64

```go
// Igual
UserFields.Age.Eq(25)

// Diferente
UserFields.Age.Ne(18)

// Maior que
UserFields.Age.Gt(18)

// Maior ou igual
UserFields.Age.Gte(18)

// Menor que
UserFields.Age.Lt(65)

// Menor ou igual
UserFields.Age.Lte(65)

// BETWEEN
UserFields.Age.Between(18, 65)

// IN
UserFields.Age.In(20, 25, 30)

// NOT IN
UserFields.Age.NotIn(0, 999)
```

#### Bool

```go
// Igual
UserFields.IsActive.Eq(true)

// Diferente
UserFields.IsActive.Ne(false)

// IS NULL
UserFields.IsActive.IsNull()

// IS NOT NULL
UserFields.IsActive.IsNotNull()
```

### Order By

```go
// ASC
users, err := genus.Table[User](db).
    OrderByAsc("name").
    Find(ctx)

// DESC
users, err := genus.Table[User](db).
    OrderByDesc("created_at").
    Find(ctx)

// Múltiplos
users, err := genus.Table[User](db).
    OrderByDesc("age").
    OrderByAsc("name").
    Find(ctx)
```

### Limit e Offset (Paginação)

```go
// Primeira página (10 itens)
users, err := genus.Table[User](db).
    OrderByAsc("id").
    Limit(10).
    Offset(0).
    Find(ctx)

// Segunda página
users, err := genus.Table[User](db).
    OrderByAsc("id").
    Limit(10).
    Offset(10).
    Find(ctx)
```

### First (Buscar Apenas Um)

```go
user, err := genus.Table[User](db).
    Where(UserFields.Email.Eq("alice@example.com")).
    First(ctx)

if err != nil {
    if err.Error() == "no rows found" {
        fmt.Println("User not found")
    } else {
        log.Fatal(err)
    }
}
```

### Count

```go
count, err := genus.Table[User](db).
    Where(UserFields.IsActive.Eq(true)).
    Count(ctx)

fmt.Printf("Active users: %d\n", count)
```

### Select (Colunas Específicas)

```go
users, err := genus.Table[User](db).
    Select("id", "name", "email").
    Find(ctx)

// Nota: Campos não selecionados terão valores zero
```

## Transações

### Transação Básica

```go
err := db.DB().WithTx(ctx, func(txDB *core.DB) error {
    // Criar usuário
    user := &User{Name: "Alice", Email: "alice@example.com"}
    if err := txDB.Create(ctx, user); err != nil {
        return err // Rollback automático
    }

    // Criar outro registro relacionado
    profile := &Profile{UserID: user.ID, Bio: "Hello"}
    if err := txDB.Create(ctx, profile); err != nil {
        return err // Rollback automático
    }

    return nil // Commit automático
})

if err != nil {
    log.Fatal(err)
}
```

### Transação com Queries

```go
err := db.DB().WithTx(ctx, func(txDB *core.DB) error {
    // Query dentro da transação
    users, err := genus.Table[User](&genus.Genus{db: txDB}).
        Where(UserFields.IsActive.Eq(true)).
        Find(ctx)

    if err != nil {
        return err
    }

    // Processar users...
    for _, user := range users {
        user.IsActive = false
        if err := txDB.Update(ctx, &user); err != nil {
            return err
        }
    }

    return nil
})
```

## Exemplos Avançados

### Busca com Paginação Helper

```go
func GetUsersPaginated(db *genus.Genus, page, pageSize int) ([]User, error) {
    ctx := context.Background()
    offset := (page - 1) * pageSize

    return genus.Table[User](db).
        OrderByDesc("created_at").
        Limit(pageSize).
        Offset(offset).
        Find(ctx)
}

// Uso
users, err := GetUsersPaginated(db, 1, 20) // Página 1, 20 itens
```

### Busca com Filtros Dinâmicos

```go
type UserFilter struct {
    Name     *string
    MinAge   *int
    IsActive *bool
}

func FindUsersWithFilter(db *genus.Genus, filter UserFilter) ([]User, error) {
    ctx := context.Background()
    builder := genus.Table[User](db)

    conditions := []interface{}{}

    if filter.Name != nil {
        conditions = append(conditions, UserFields.Name.Like("%"+*filter.Name+"%"))
    }

    if filter.MinAge != nil {
        conditions = append(conditions, UserFields.Age.Gte(*filter.MinAge))
    }

    if filter.IsActive != nil {
        conditions = append(conditions, UserFields.IsActive.Eq(*filter.IsActive))
    }

    for _, cond := range conditions {
        builder = builder.Where(cond)
    }

    return builder.Find(ctx)
}

// Uso
name := "Alice"
minAge := 18
active := true

users, err := FindUsersWithFilter(db, UserFilter{
    Name:     &name,
    MinAge:   &minAge,
    IsActive: &active,
})
```

### Repository Pattern

```go
type UserRepository struct {
    db *genus.Genus
}

func NewUserRepository(db *genus.Genus) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    return genus.Table[User](r.db).
        Where(UserFields.Email.Eq(email)).
        First(ctx)
}

func (r *UserRepository) FindActive(ctx context.Context) ([]User, error) {
    return genus.Table[User](r.db).
        Where(UserFields.IsActive.Eq(true)).
        OrderByDesc("created_at").
        Find(ctx)
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    return r.db.DB().Create(ctx, user)
}

// Uso
repo := NewUserRepository(db)
users, err := repo.FindActive(ctx)
```

## Debugging

### Ver SQL Gerado

Para debugar queries, você pode inspecionar o código-fonte do builder ou adicionar logging:

```go
// TODO: Adicionar suporte a logging de queries em versão futura
// db.DB().Logger = customLogger
```

## Best Practices

1. **Sempre use context**: Todas as operações aceitam `context.Context`
2. **Defina campos tipados**: Crie `XFields` para cada modelo
3. **Use transações**: Para operações que modificam múltiplos registros
4. **Repository pattern**: Organize queries em repositories
5. **Validação em hooks**: Use `BeforeCreate` para validações

## Limitações Atuais

- Não suporta relações (HasMany, BelongsTo) ainda
- Não tem migrations automáticas
- Não tem soft deletes built-in
- Não tem eager loading/preloading

Estas features estão no roadmap!
