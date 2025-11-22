# Entendendo Go Generics no Genus

Este documento explica **de forma didática** como o Genus usa Go Generics para alcançar type-safety completa.

## O Problema: ORMs Tradicionais

### GORM e outros ORMs antigos

```go
// GORM - não é type-safe
var users []User
db.Where("name = ?", "Alice").Find(&users)

// ❌ Problemas:
// 1. "name" é uma string mágica (typo não é detectado)
// 2. "Alice" poderia ser qualquer tipo (não verificado)
// 3. Precisa passar &users (ponteiro para slice)
// 4. Usa muita reflection (lento)
```

## A Solução: Go Generics (Go 1.18+)

### 1. O Básico de Generics

Generics permitem escrever código que funciona com **qualquer tipo**, mas de forma **type-safe**.

#### Exemplo Simples

```go
// Antes de Generics (Go < 1.18)
func PrintInt(value int) {
    fmt.Println(value)
}

func PrintString(value string) {
    fmt.Println(value)
}

// Com Generics (Go 1.18+)
func Print[T any](value T) {
    fmt.Println(value)
}

// Uso
Print[int](42)        // T = int
Print[string]("foo")  // T = string
Print(42)             // Inferência: T = int
```

O `[T any]` significa: "T é um tipo genérico que pode ser qualquer coisa".

### 2. Como o Genus Usa Generics

#### A. Query Builder Genérico

```go
// query/builder.go
type Builder[T any] struct {
    tableName  string
    conditions []Condition
    // ...
}

func NewBuilder[T any](tableName string) *Builder[T] {
    return &Builder[T]{
        tableName: tableName,
    }
}

func (b *Builder[T]) Find(ctx context.Context) ([]T, error) {
    // ... executa query SQL ...
    var results []T
    // ... scan rows into results ...
    return results, nil  // ✅ Retorna []T diretamente!
}
```

**Como isso funciona:**

1. Quando você escreve `genus.Table[User](db)`:
   - `T` é substituído por `User` em **compile-time**
   - Você recebe um `*Builder[User]`

2. Quando você chama `.Find(ctx)`:
   - O retorno é `[]User` (não `[]interface{}`)
   - O compilador **garante** que é type-safe!

#### B. A Função Table

```go
// genus.go
func Table[T any](g *Genus) *query.Builder[T] {
    var model T
    tableName := getTableName(model)
    return query.NewBuilder[T](g.db.Executor(), g.db.Dialect(), tableName)
}
```

**Fluxo completo:**

```go
// 1. Você escreve:
users, err := genus.Table[User](db).Find(ctx)

// 2. O compilador substitui T por User:
users, err := genus.Table[User](db).Find(ctx)
//                     ^^^^ T = User
//                          retorna *Builder[User]

// 3. Find retorna []User:
users, err := builder.Find(ctx)  // []User, error
//            ^^^^ retorna []User (não []interface{})
```

**Vantagem:** O compilador sabe que `users` é `[]User`, não precisa de type assertion!

### 3. Campos Tipados: O Design Mais Importante

#### O Problema

Como garantir que você não pode fazer isso?

```go
// ❌ Queremos evitar isso:
db.Where("age", "not a number") // String comparada com int!
```

#### A Solução: Tipos Específicos para Cada Campo

```go
// query/field.go

// Tipo específico para campos string
type StringField struct {
    column string
}

func (f StringField) Eq(value string) Condition {
    //                    ^^^^^^ DEVE ser string!
    return Condition{
        Field:    f.column,
        Operator: OpEq,
        Value:    value,
    }
}

// Tipo específico para campos int
type IntField struct {
    column string
}

func (f IntField) Eq(value int) Condition {
    //                   ^^^ DEVE ser int!
    return Condition{
        Field:    f.column,
        Operator: OpEq,
        Value:    value,
    }
}
```

**Como usar:**

```go
var UserFields = struct {
    Name query.StringField
    Age  query.IntField
}{
    Name: query.NewStringField("name"),
    Age:  query.NewIntField("age"),
}

// ✅ Type-safe!
UserFields.Name.Eq("Alice")  // OK: string com string
UserFields.Age.Eq(25)        // OK: int com int

// ❌ Erro de compilação!
UserFields.Age.Eq("25")      // ERRO: expected int, got string
UserFields.Name.Gt(10)       // ERRO: StringField não tem método Gt
```

**Magia:** O compilador **garante** que você só pode comparar tipos compatíveis!

### 4. Interfaces Genéricas

#### Definindo Comportamentos Type-Safe

```go
// query/field.go

// Interface genérica: T é o tipo do valor
type Comparador[T any] interface {
    Eq(value T) Condition
    Ne(value T) Condition
    In(values ...T) Condition
}

// Interface para tipos ordenáveis
type ComparadorOrdenavel[T any] interface {
    Comparador[T]
    Gt(value T) Condition
    Lt(value T) Condition
}
```

**Implementação:**

```go
// StringField implementa Comparador[string]
type StringField struct {
    column string
}

func (f StringField) Eq(value string) Condition {
    // ...
}

func (f StringField) Like(pattern string) Condition {
    // Método específico para strings!
    // ...
}

// IntField implementa ComparadorOrdenavel[int]
type IntField struct {
    column string
}

func (f IntField) Eq(value int) Condition {
    // ...
}

func (f IntField) Gt(value int) Condition {
    // Método para tipos ordenáveis!
    // ...
}
```

**Resultado:**

```go
UserFields.Age.Gt(18)        // ✅ OK: int é ordenável
UserFields.Name.Like("%a%")  // ✅ OK: string tem Like
UserFields.Name.Gt("abc")    // ❌ ERRO: string não tem Gt
UserFields.Age.Like("%")     // ❌ ERRO: int não tem Like
```

### 5. Por Que Não Precisamos de `*[]T`?

#### O Problema em GORM

```go
// GORM requer ponteiro para slice
var users []User
db.Find(&users)  // Precisa de &
```

Por quê? Porque GORM usa reflection para **modificar** o slice que você passou.

#### Nossa Solução

```go
// Genus retorna o slice diretamente
users, err := genus.Table[User](db).Find(ctx)
```

**Como funciona internamente:**

```go
func (b *Builder[T]) Find(ctx context.Context) ([]T, error) {
    // 1. Criamos o slice DENTRO da função
    var results []T

    // 2. Executamos a query
    rows, _ := b.executor.QueryContext(ctx, query, args...)

    // 3. Preenchemos o slice
    for rows.Next() {
        var item T
        scanStruct(rows, &item)  // Usa reflection AQUI (isolado)
        results = append(results, item)
    }

    // 4. Retornamos o slice
    return results, nil  // ✅ Slice já está pronto!
}
```

**Vantagens:**

1. API mais limpa (não precisa de `&`)
2. Impossível passar tipo errado (type-safe)
3. Reflection isolada (só no scanning)

### 6. Comparação Lado a Lado

#### GORM (Tradicional)

```go
// 1. Definir modelo
type User struct {
    ID   uint
    Name string
    Age  int
}

// 2. Query
var users []User
db.Where("name = ?", "Alice").  // ❌ String mágica
   Where("age > ?", 18).         // ❌ Não type-safe
   Find(&users)                  // ❌ Precisa de &

// 3. Possíveis erros em runtime:
db.Where("nme = ?", "Alice")    // ❌ Typo - só detectado em runtime!
db.Where("age > ?", "texto")    // ❌ Tipo errado - erro em runtime!
```

#### Genus (Type-Safe)

```go
// 1. Definir modelo
type User struct {
    core.Model
    Name string `db:"name"`
    Age  int    `db:"age"`
}

// 2. Definir campos (uma vez)
var UserFields = struct {
    Name query.StringField
    Age  query.IntField
}{
    Name: query.NewStringField("name"),
    Age:  query.NewIntField("age"),
}

// 3. Query
users, err := genus.Table[User](db).
    Where(UserFields.Name.Eq("Alice")).   // ✅ Type-safe
    Where(UserFields.Age.Gt(18)).         // ✅ Type-safe
    Find(ctx)                             // ✅ Retorna []User

// 4. Erros detectados em compile-time:
UserFields.Nme.Eq("Alice")     // ❌ ERRO DE COMPILAÇÃO: Nme não existe
UserFields.Age.Eq("texto")     // ❌ ERRO DE COMPILAÇÃO: expected int
UserFields.Name.Gt(10)         // ❌ ERRO DE COMPILAÇÃO: método não existe
```

### 7. O Trade-off: Verbosidade vs Safety

#### O Custo

Você precisa definir campos manualmente:

```go
var UserFields = struct {
    Name query.StringField
    Age  query.IntField
}{
    Name: query.NewStringField("name"),
    Age:  query.NewIntField("age"),
}
```

Isso é mais verboso que GORM.

#### O Benefício

**Todos** os erros são detectados em **compile-time**:

```go
// ✅ Autocomplete funciona perfeitamente
UserFields.Name.  // IDE mostra: Eq, Ne, Like, NotLike, In, NotIn, ...

// ✅ Refactoring é seguro
// Renomear Age para BirthYear?
// O compilador encontra TODOS os usos!

// ✅ Impossível fazer queries inválidas
UserFields.Age.Eq("string")  // Não compila!
```

#### A Solução Futura: Code Generation

```bash
# Futuro
$ genus generate ./models

# Gera automaticamente:
# models/user_fields.gen.go
var UserFields = struct {
    // ... gerado automaticamente ...
}{
    // ...
}
```

### 8. Quando Ainda Usamos Reflection

Genus **minimiza** reflection, mas ainda usa em dois lugares:

#### A. Scanning (query/scanner.go)

```go
func scanStruct(rows *sql.Rows, dest interface{}) error {
    // Usa reflection para mapear colunas do DB para campos da struct
    destValue := reflect.ValueOf(dest).Elem()
    // ...
}
```

**Por quê?** `database/sql.Rows` não é tipado. Precisamos usar reflection para descobrir os campos da struct.

#### B. CRUD (core/db.go)

```go
func (db *DB) Create(ctx context.Context, model interface{}) error {
    // Usa reflection para extrair campos
    columns, values := getColumnsAndValues(model)
    // ...
}
```

**Por quê?** Create/Update/Delete não sabem antecipadamente quais campos existem.

**Nota:** Podemos eliminar isso com generics também, mas seria mais complexo.

### 9. Resumo: Por Que Generics São Importantes

| Aspecto | Sem Generics | Com Generics |
|---------|--------------|--------------|
| **Type Safety** | Runtime | Compile-time |
| **Autocomplete** | Limitado | Completo |
| **Refactoring** | Arriscado | Seguro |
| **Performance** | Reflection pesada | Reflection mínima |
| **Erros** | Runtime panics | Compile errors |
| **API** | `Find(&users)` | `Find(ctx)` retorna `[]User` |

### 10. Exercício Prático

Tente escrever código inválido e veja o compilador reclamar:

```go
// ❌ Todos esses erros são detectados pelo compilador:

// 1. Tipo errado
UserFields.Age.Eq("not a number")

// 2. Método inexistente
UserFields.Name.Gt("abc")

// 3. Campo inexistente
UserFields.NonExistent.Eq("foo")

// 4. Operador errado para o tipo
UserFields.IsActive.Between(true, false)

// 5. Retorno com tipo errado
var users []Product = genus.Table[User](db).Find(ctx)
```

**Todos esses erros seriam runtime panics em GORM!**

## Conclusão

Go Generics permitem que o Genus seja:

1. **Type-safe**: Erros detectados em compile-time
2. **Performático**: Menos reflection
3. **Fácil de usar**: API limpa e intuitiva
4. **Seguro para refactoring**: Compilador ajuda

O custo é verbosidade na definição de campos, mas isso pode ser resolvido com code generation.

**Genus mostra que Go Generics podem criar ORMs modernos, seguros e rápidos!** 🚀
