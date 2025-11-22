# Changelog

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
e este projeto adere ao [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2024-01-XX

### Corrigido

#### 1. **Bug Crítico: Scanning com core.Model Embedded**
- **Problema:** Ao fazer scan de resultados SQL para structs com `core.Model` embedded, ocorria erro pois o código tentava escanear para o campo embedded inteiro ao invés dos campos individuais (ID, CreatedAt, UpdatedAt)
- **Solução:** Implementado sistema de `fieldPath` em `query/scanner.go` que navega corretamente através de structs embedded usando um caminho de índices
- **Impacto:** Agora todas as queries com modelos que usam `core.Model` funcionam corretamente
- **Arquivo:** `query/scanner.go`

#### 2. **Bug: AutoMigrate com SQLite (Sintaxe MySQL Incorreta)**
- **Problema:** `migrate.AutoMigrate()` gerava SQL com sintaxe `AUTO_INCREMENT` do MySQL, que SQLite não suporta
- **Solução:** Implementado detecção de dialeto baseada em `Placeholder()` e `QuoteIdentifier()`:
  - PostgreSQL: usa `SERIAL`
  - MySQL: usa `INTEGER AUTO_INCREMENT`
  - SQLite: usa `INTEGER` (auto-increment automático para INTEGER PRIMARY KEY)
- **Impacto:** AutoMigrate agora funciona corretamente em SQLite
- **Arquivo:** `migrate/auto.go`

#### 3. **Bug: createMigrationsTable Usando Tipos Genéricos**
- **Problema:** Tabela de migrations usava tipos genéricos (VARCHAR, TIMESTAMP) que podem não funcionar em todos os dialetos
- **Solução:** Implementado detecção de dialeto e uso de tipos específicos:
  - PostgreSQL: `BIGINT`, `VARCHAR(255)`, `TIMESTAMP`
  - MySQL: `BIGINT`, `VARCHAR(255)`, `DATETIME`
  - SQLite: `INTEGER`, `TEXT`, `DATETIME`
- **Impacto:** Sistema de migrations funciona corretamente em todos os dialetos suportados
- **Arquivo:** `migrate/migrate.go`

### Adicionado

#### 1. **core.ErrValidation**
- Adicionado erro `ErrValidation` que estava faltando e era usado nos exemplos
- **Arquivo:** `core/interfaces.go`

#### 2. **Float64Field**
- Adicionado tipo `Float64Field` para campos float64 não-nullable
- Já existia `OptionalFloat64Field` mas faltava a versão não-opcional
- Suporta todos os operadores: Eq, Ne, Gt, Gte, Lt, Lte, In, NotIn, Between, IsNull, IsNotNull
- **Arquivo:** `query/field.go`

#### 3. **Dependências SQL Drivers**
- Adicionadas dependências dos drivers SQL oficiais:
  - `github.com/lib/pq` (PostgreSQL)
  - `github.com/go-sql-driver/mysql` (MySQL)
  - `github.com/mattn/go-sqlite3` (SQLite)
- **Arquivo:** `go.mod`

### Corrigido - Exemplos

#### 1. **Atualizados Exemplos para API Correta**
- Corrigido uso de `genus.New()` para `genus.NewWithLogger()`
- Corrigido chamadas de CRUD: `g.Create()` → `g.DB().Create()`
- Corrigido Table builder: `g.Table[T]()` → `genus.Table[T](g)`
- **Arquivos:** `examples/optional/main.go`, `examples/migrations/main.go`, `examples/multi-database/main.go`

#### 2. **Corrigidos fmt.Println com Newlines Redundantes**
- Substituído `fmt.Println("text\n")` por `fmt.Println("text")` seguido de `fmt.Println()`
- Corrige warning do linter: "fmt.Println arg list ends with redundant newline"
- **Arquivos:** Todos os exemplos

## [1.0.0] - 2024-01-XX

### Adicionado - Versão 1.x (Usabilidade - Performance e Composição)

#### 1. Tipos Opcionais Genéricos (`Optional[T]`)

**Motivação:** Resolver a dor da manipulação de `sql.Null*` e ponteiros em JSON com uma API limpa e unificada.

**Funcionalidades:**
- Tipo genérico `core.Optional[T]` para valores nullable
- Suporte completo a JSON marshaling/unmarshaling (serializa como `null` quando vazio)
- Implementa `sql.Scanner` e `driver.Valuer` para integração com database/sql
- API funcional: `Map`, `FlatMap`, `Filter`, `IfPresent`, `IfPresentOrElse`
- Funções helper: `Some()`, `None()`, `FromPtr()`
- Métodos de acesso: `Get()`, `GetOrDefault()`, `GetOrZero()`, `Ptr()`
- Conversões automáticas para tipos primitivos (string, int, int64, bool, float64)

**Exemplo:**
```go
type User struct {
    core.Model
    Name  string                `db:"name"`
    Email core.Optional[string] `db:"email"`  // Pode ser NULL
    Age   core.Optional[int]    `db:"age"`    // Pode ser NULL
}

// Criar valores
email := core.Some("user@example.com")
age := core.None[int]()

// Usar
if email.IsPresent() {
    fmt.Println(email.Get())
}
userAge := age.GetOrDefault(18)
```

**Arquivos:**
- `core/optional.go` - Implementação completa do tipo Optional[T]

**Supera:** Todos os ORMs Go existentes - primeira implementação completa de Optional[T] genérico para Go

---

#### 2. Campos Opcionais Tipados

**Motivação:** Permitir queries type-safe em campos nullable.

**Funcionalidades:**
- `OptionalStringField` - Campo string opcional
- `OptionalIntField` - Campo int opcional
- `OptionalInt64Field` - Campo int64 opcional
- `OptionalBoolField` - Campo bool opcional
- `OptionalFloat64Field` - Campo float64 opcional
- Todos os campos suportam operadores apropriados (`Eq`, `Ne`, `Gt`, `Like`, `IsNull`, `IsNotNull`, etc.)

**Exemplo:**
```go
var UserFields = struct {
    Name  query.StringField
    Email query.OptionalStringField  // Campo opcional
    Age   query.OptionalIntField     // Campo opcional
}{
    Name:  query.NewStringField("name"),
    Email: query.NewOptionalStringField("email"),
    Age:   query.NewOptionalIntField("age"),
}

// Query em campos opcionais
users, _ := genus.Table[User]().
    Where(UserFields.Email.IsNotNull()).
    Where(UserFields.Age.Gt(18)).
    Find(ctx)
```

**Arquivos:**
- `query/field.go` - Adicionados campos opcionais (linhas 362-794)

**Supera:** GORM, Squirrel - campos opcionais totalmente tipados

---

#### 3. Query Builder Imutável

**Motivação:** Permitir composição segura de queries dinâmicas sem efeitos colaterais.

**Funcionalidades:**
- Todos os métodos do Builder retornam uma nova instância
- Método `clone()` interno para cópia profunda do estado
- Thread-safe por design
- Permite reutilização de queries base sem mutação

**Exemplo:**
```go
// Base query não é modificada
baseQuery := genus.Table[User]().Where(UserFields.Active.Eq(true))

// Composição segura
adults := baseQuery.Where(UserFields.Age.Gte(18))
minors := baseQuery.Where(UserFields.Age.Lt(18))

// baseQuery permanece inalterada!
// Cada query é completamente independente
```

**Impacto:**
- Antes: `baseQuery.Where()` modificava o objeto original
- Depois: `baseQuery.Where()` retorna uma nova query, original intocado

**Arquivos:**
- `query/builder.go` - Adicionado método `clone()` e modificados todos os métodos de building (linhas 42-132)

**Supera:** Squirrel - composição type-safe e imutável

---

#### 4. CLI de Code Generation (`genus generate`)

**Motivação:** Eliminar boilerplate manual e garantir sincronização automática entre structs e campos tipados.

**Funcionalidades:**
- CLI completo com comandos `generate`, `version`, `help`
- Parser de AST Go para extrair structs e tags `db`
- Geração automática de arquivos `*_fields.gen.go`
- Detecção automática de tipos `Optional[T]`
- Mapeamento inteligente de tipos Go para tipos de campo query
- Flags: `-o` (output dir), `-p` (package name), `-h` (help)

**Uso:**
```bash
# Instalar
go install github.com/GabrielOnRails/genus/cmd/genus@latest

# Gerar campos
genus generate ./models

# Com flags
genus generate -o ./generated -p mypackage ./models
```

**Entrada (struct):**
```go
type User struct {
    core.Model
    Name  string                `db:"name"`
    Email core.Optional[string] `db:"email"`
    Age   core.Optional[int]    `db:"age"`
}
```

**Saída (gerada automaticamente):**
```go
// user_fields.gen.go
var UserFields = struct {
    ID    query.Int64Field
    Name  query.StringField
    Email query.OptionalStringField
    Age   query.OptionalIntField
}{
    ID:    query.NewInt64Field("id"),
    Name:  query.NewStringField("name"),
    Email: query.NewOptionalStringField("email"),
    Age:   query.NewOptionalIntField("age"),
}
```

**Arquivos:**
- `cmd/genus/main.go` - CLI principal com comandos
- `codegen/generator.go` - Lógica de geração (parser AST, extração de structs, mapeamento de tipos)
- `codegen/template.go` - Template de código gerado

**Supera:** GORM - remove dependência excessiva de runtime reflection ao gerar metadados de coluna em compile-time

---

### Mudanças Técnicas

#### Performance
- **Zero reflection em queries:** Campos tipados gerados eliminam necessidade de reflection para descobrir metadados de coluna
- **Builder imutável:** Clone otimizado com cópia profunda apenas de slices necessários
- **Optional[T]:** Implementação eficiente com conversões diretas para tipos primitivos

#### Arquitetura
- Novo pacote `codegen` para geração de código
- Novo pacote `cmd/genus` para CLI
- Expansão de `core` com tipo `Optional[T]`
- Expansão de `query` com campos opcionais

#### Compatibilidade
- **Breaking change:** Query builder agora é imutável
  - Migração: Nenhuma mudança necessária no código do usuário (API permanece a mesma)
  - Impacto: Queries agora são thread-safe e podem ser reutilizadas

---

### Exemplos e Documentação

#### Novos Exemplos
- `examples/optional/main.go` - Demonstração completa de Optional[T]
- `examples/codegen/models/user.go` - Modelos para code generation
- `examples/codegen/README.md` - Tutorial de code generation

#### Documentação Atualizada
- `README.md` - Adicionadas seções sobre Optional[T], Code Generation e Query Builder Imutável
- `README.md` - Tabela de comparação expandida (GORM, Ent, sqlboiler, Squirrel)
- `examples/codegen/README.md` - Guia completo de uso do CLI

---

### Comparação de Performance (vs Competidores)

| Métrica | GORM | sqlboiler | Squirrel | **Genus 1.x** |
|---------|------|-----------|----------|---------------|
| Reflection em queries | Alto | Zero | N/A | Zero (após codegen) |
| Type-safety | Baixo | Alto | Baixo | Alto |
| Imutabilidade | Não | Não | Não | Sim |
| Tipos opcionais | `sql.Null*` | `null.*` | Manual | `Optional[T]` |
| Code generation | Não | Sim (schemas) | Não | Sim (fields) |

---

## [0.1.0] - 2024-01-XX

### Adicionado - Versão Inicial

#### Core Features
- Query builder genérico com suporte a Go Generics
- Campos tipados (StringField, IntField, Int64Field, BoolField)
- Operadores type-safe (Eq, Ne, Gt, Like, etc.)
- Suporte a condições complexas (AND/OR)
- CRUD operations (Create, Update, Delete)
- SQL logging automático com performance monitoring
- Suporte a PostgreSQL (dialect)
- Transaction support
- Hook system (BeforeCreate, AfterFind)
- Context-aware operations
- Direct slice returns (`[]T`)

#### Packages
- `core` - Tipos base, DB, Logger, Interfaces
- `query` - Query builder, Fields, Conditions
- `dialects/postgres` - PostgreSQL dialect

#### Examples
- `examples/basic` - Exemplo completo de todas as features
- `examples/logging` - Configuração de logging customizado
- `examples/testing` - Padrões de teste

---

## Próximas Versões

### [2.0.0] - Planejado

#### Relational Features
- HasMany, BelongsTo, ManyToMany relationships
- Eager loading / Preloading
- Join support

#### Database Support
- MySQL dialect
- SQLite dialect

#### Advanced Features
- Migrations automáticas
- Soft deletes
- Advanced hooks (AfterCreate, BeforeUpdate, etc.)
- Query caching
- Connection pooling configuration

---

[1.0.0]: https://github.com/GabrielOnRails/genus/releases/tag/v1.0.0
[0.1.0]: https://github.com/GabrielOnRails/genus/releases/tag/v0.1.0
