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

### 6. SQL Logging (Transparência Total)

Genus loga **automaticamente** todas as queries SQL executadas, incluindo tempo de execução:

```go
// Logging padrão (não-verbose) - habilitado automaticamente
db, _ := genus.Open("postgres", "...")
genus.Table[User](db).Where(UserFields.Age.Gt(25)).Find(ctx)
// Output: [GENUS] 2.34ms | SELECT * FROM "users" WHERE age > $1

// Logging verbose (mostra argumentos)
sqlDB, _ := sql.Open("postgres", "...")
verboseDB := genus.New(sqlDB, core.NewDefaultLogger(true))
// Output: [GENUS] 2.34ms | SELECT * FROM "users" WHERE age > $1 | args: [25]

// Desabilitar logging
silentDB := genus.New(sqlDB, &core.NoOpLogger{})

// Logger customizado (JSON, arquivo, métricas, etc)
type MyLogger struct{}
func (l *MyLogger) LogQuery(query string, args []interface{}, duration int64) {
    // Envie para seu sistema de logging
}
func (l *MyLogger) LogError(query string, args []interface{}, err error) {
    // Trate erros
}

customDB := genus.New(sqlDB, &MyLogger{})
```

**Vantagens do SQL Logging:**
- Debugging facilitado: veja exatamente qual SQL está sendo executado
- Performance monitoring: tempo de execução em cada query
- Auditoria: rastreie todas as operações no banco
- Customizável: implemente `core.Logger` para enviar logs para onde quiser

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

| Característica | GORM | Ent | sqlboiler | Squirrel | **Genus 1.x** |
|---------------|------|-----|-----------|----------|---------------|
| Type-safe queries | ❌ | ✅ | ✅ | ❌ | ✅ |
| Retorna `[]T` | ❌ | ✅ | ✅ | N/A | ✅ |
| Code generation opcional | ❌ | ✅ | ✅ | ❌ | ✅ |
| Reflection mínimo | ❌ | ✅ | ✅ | N/A | ✅ |
| Campos tipados gerados | ❌ | ✅ | ⚠️ | ❌ | ✅ |
| API fluente | ✅ | ✅ | ⚠️ | ✅ | ✅ |
| Query builder imutável | ❌ | ✅ | ❌ | ❌ | ✅ |
| Tipos Optional[T] | ❌ | ❌ | ⚠️ | ❌ | ✅ |
| SQL logging automático | ⚠️ | ⚠️ | ❌ | ❌ | ✅ |
| Performance monitoring | ❌ | ❌ | ❌ | ❌ | ✅ |
| Zero dependencies | ❌ | ❌ | ❌ | ✅ | ✅ |

**Legenda:**
- ✅ Suporte completo
- ⚠️ Suporte parcial
- ❌ Não suportado
- N/A Não aplicável

## Recursos Implementados

### 1. Suporte Multi-Database

Genus suporta PostgreSQL, MySQL e SQLite através de dialects:

```go
import (
    "github.com/gabrieldias/genus/dialects/postgres"
    "github.com/gabrieldias/genus/dialects/mysql"
    "github.com/gabrieldias/genus/dialects/sqlite"
)

// PostgreSQL
g := genus.New(db, postgres.New(), logger)

// MySQL
g := genus.New(db, mysql.New(), logger)

// SQLite
g := genus.New(db, sqlite.New(), logger)
```

### 2. Sistema de Migrations

Genus oferece migrations automáticas e manuais:

```go
import "github.com/gabrieldias/genus/migrate"

// AutoMigrate (desenvolvimento)
migrate.AutoMigrate(ctx, db, dialect, User{}, Product{})

// Manual Migrations (produção)
migrator := migrate.New(db, dialect, logger, migrate.Config{})
migrator.Register(migrate.Migration{
    Version: 1,
    Name: "create_users_table",
    Up: func(ctx, db, dialect) error { /* ... */ },
    Down: func(ctx, db, dialect) error { /* ... */ },
})
migrator.Up(ctx)
```

**CLI de Migrations:**
```bash
genus migrate create add_users_table  # Criar migration
genus migrate up                      # Aplicar migrations
genus migrate down                    # Reverter última migration
genus migrate status                  # Ver status
```

## Versão 1.x - Recursos Principais

### 3. Optional[T] - Tipos Opcionais Genéricos

Genus agora oferece suporte completo para campos nullable com uma API limpa e type-safe:

```go
type User struct {
    core.Model
    Name  string                `db:"name"`
    Email core.Optional[string] `db:"email"`  // Pode ser NULL
    Age   core.Optional[int]    `db:"age"`    // Pode ser NULL
}

// Criar valores Optional
email := core.Some("user@example.com")  // Valor presente
age := core.None[int]()                 // Valor ausente

// Verificar e usar
if email.IsPresent() {
    fmt.Println(email.Get())
}

// Obter com valor padrão
userAge := age.GetOrDefault(18)

// Operações funcionais
upperEmail := core.Map(email, strings.ToUpper)
filtered := age.Filter(func(a int) bool { return a > 18 })
```

**Vantagens do Optional[T]:**
- ✅ API consistente e limpa (sem `sql.Null*` ou ponteiros)
- ✅ Suporte automático para JSON marshaling/unmarshaling
- ✅ Implementa `sql.Scanner` e `driver.Valuer`
- ✅ Operações funcionais (Map, Filter, FlatMap)
- ✅ Type-safe em tempo de compilação

### 4. Code Generation CLI

Gere campos tipados automaticamente a partir de structs Go:

```bash
# Instalar CLI
go install github.com/gabrieldias/genus/cmd/genus@latest

# Gerar campos tipados
genus generate ./models

# Resultado: cria arquivos *_fields.gen.go
```

**Antes (manual):**
```go
var UserFields = struct {
    Name  query.StringField
    Email query.OptionalStringField
    Age   query.OptionalIntField
}{
    Name:  query.NewStringField("name"),
    Email: query.NewOptionalStringField("email"),
    Age:   query.NewOptionalIntField("age"),
}
```

**Depois (gerado automaticamente):**
```bash
genus generate ./models
# Cria user_fields.gen.go com todos os campos tipados!
```

**Vantagens:**
- ✅ Zero boilerplate manual
- ✅ Campos sempre sincronizados com structs
- ✅ Detecta automaticamente tipos Optional[T]
- ✅ Integração fácil com CI/CD

### 5. Query Builder Imutável

O query builder agora é completamente imutável, permitindo composição segura:

```go
// Base query não é modificada
baseQuery := genus.Table[User]().Where(UserFields.Active.Eq(true))

// Composição segura - cada query é independente
adults := baseQuery.Where(UserFields.Age.Gte(18))
minors := baseQuery.Where(UserFields.Age.Lt(18))

// baseQuery permanece inalterada!
// adults e minors são queries completamente separadas
```

**Vantagens:**
- ✅ Queries podem ser reutilizadas sem efeitos colaterais
- ✅ Thread-safe por design
- ✅ Facilita testes e composição
- ✅ Supera Squirrel em segurança de tipos

## Roadmap

### Versão 1.x ✅ (Implementado)
- [x] Optional[T] - Tipos opcionais genéricos
- [x] Code generation CLI (genus generate)
- [x] Query builder imutável
- [x] Campos opcionais tipados (OptionalStringField, etc.)
- [x] Suporte para MySQL e SQLite
- [x] Migrations automáticas (AutoMigrate + Manual)
- [x] CLI de migrations (genus migrate)

### Versão 2.x (Planejado)
- [ ] Relações (HasMany, BelongsTo, ManyToMany)
- [ ] Eager loading / Preloading
- [ ] Join support
- [ ] Hooks avançados (AfterCreate, BeforeUpdate, etc.)
- [ ] Soft deletes
- [ ] Query caching
- [ ] Connection pooling configuration

## Exemplos

O projeto inclui vários exemplos práticos:

- **`examples/basic/main.go`** - Exemplo completo com todas as funcionalidades básicas
- **`examples/optional/main.go`** - Demonstração do uso de Optional[T] com banco de dados
- **`examples/codegen/`** - Exemplo de code generation com genus CLI
- **`examples/multi-database/`** - Uso com PostgreSQL, MySQL e SQLite
- **`examples/migrations/`** - Sistema completo de migrations (AutoMigrate + Manual)
- **`examples/logging/main.go`** - Configuração de logging customizado
- **`examples/testing/`** - Padrões de teste com repository pattern

### Executar Exemplos

```bash
# Exemplo básico
go run examples/basic/main.go

# Exemplo de Optional[T]
go run examples/optional/main.go

# Code generation
cd examples/codegen/models
genus generate .

# Multi-database
go run examples/multi-database/main.go

# Migrations
go run examples/migrations/main.go
```

## Desenvolvimento

### Setup Inicial

Após clonar o repositório, execute o script de setup para configurar os git hooks:

```bash
./scripts/setup-hooks.sh
```

Este script instala hooks que validam mensagens de commit, garantindo consistência no histórico do projeto.

### Git Hooks

O projeto utiliza os seguintes hooks:

- **commit-msg**: Valida mensagens de commit, bloqueando referências a ferramentas de IA externas

Para reinstalar os hooks: `./scripts/setup-hooks.sh`

## Licença

MIT

## Contribuindo

Contribuições são bem-vindas! Por favor, abra uma issue ou PR.

### Processo de Contribuição

1. Clone o repositório
2. Execute `./scripts/setup-hooks.sh` para configurar os hooks
3. Crie uma branch para sua feature: `git checkout -b feature/minha-feature`
4. Faça suas mudanças e commits (os hooks validarão automaticamente)
5. Abra um Pull Request
