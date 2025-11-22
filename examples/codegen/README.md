# Code Generation Example

Este exemplo demonstra como usar o CLI `genus generate` para gerar campos tipados automaticamente a partir de structs Go.

## Estrutura

```
codegen/
├── models/
│   ├── user.go           # Structs com tags db
│   └── *_fields.gen.go   # Arquivos gerados (após rodar genus generate)
└── README.md             # Este arquivo
```

## Como Usar

### 1. Instalar o CLI

Primeiro, compile e instale o CLI do genus:

```bash
cd /Users/gabrieldias/PESSOAL/GENUS
go install ./cmd/genus
```

### 2. Gerar Campos Tipados

Execute o comando de geração no diretório de models:

```bash
# Da raiz do projeto
genus generate ./examples/codegen/models

# Ou do diretório models
cd examples/codegen/models
genus generate .
```

### 3. Verificar Arquivos Gerados

O comando criará arquivos `*_fields.gen.go` com campos tipados:

```go
// user_fields.gen.go
package models

import (
	"github.com/gabrieldias/genus/query"
	"github.com/gabrieldias/genus/core"
)

var UserFields = struct {
	ID        query.Int64Field
	Name      query.StringField
	Email     query.StringField
	Username  query.StringField
	Bio       query.OptionalStringField
	Age       query.OptionalIntField
	Verified  query.BoolField
	Premium   query.OptionalBoolField
	LastLogin query.OptionalInt64Field
	Rating    query.OptionalFloat64Field
	CreatedAt query.StringField
	UpdatedAt query.StringField
}{
	ID:        query.NewInt64Field("id"),
	Name:      query.NewStringField("name"),
	Email:     query.NewStringField("email"),
	// ... etc
}
```

## Uso dos Campos Gerados

Depois de gerar os campos, você pode usá-los em queries:

```go
package main

import (
	"context"
	"fmt"

	"github.com/gabrieldias/genus"
	"github.com/gabrieldias/genus/examples/codegen/models"
)

func main() {
	// Setup database connection
	g := genus.Open("postgres://...")

	// Query usando campos gerados
	users, err := g.Table[models.User]().
		Where(models.UserFields.Verified.Eq(true)).
		Where(models.UserFields.Age.Gt(18)).
		OrderByDesc(models.UserFields.CreatedAt.ColumnName()).
		Limit(10).
		Find(context.Background())

	if err != nil {
		panic(err)
	}

	for _, user := range users {
		fmt.Printf("User: %s (%s)\n", user.Name, user.Email)

		// Usar Optional fields
		if user.Bio.IsPresent() {
			fmt.Printf("  Bio: %s\n", user.Bio.Get())
		}

		if user.Age.IsPresent() {
			fmt.Printf("  Age: %d\n", user.Age.Get())
		}
	}
}
```

## Vantagens do Code Generation

### 1. **Type-Safety**
- Erros de digitação são detectados em tempo de compilação
- IDE fornece autocompleção para nomes de campos

### 2. **Zero Reflection em Queries**
- Metadados de coluna são gerados em tempo de compilação
- Performance superior ao GORM

### 3. **Suporte a Optional[T]**
- Campos nullable são mapeados automaticamente para OptionalXXXField
- API consistente para trabalhar com valores NULL

### 4. **Manutenção Automática**
- Adicione um campo na struct
- Rode `genus generate`
- Campos tipados são atualizados automaticamente

## Opções do CLI

```bash
# Gerar do diretório atual
genus generate

# Gerar de um diretório específico
genus generate ./models

# Especificar diretório de saída
genus generate -o ./generated ./models

# Especificar nome do pacote
genus generate -p mypackage ./models

# Ajuda
genus generate --help
```

## Tags Suportadas

O gerador processa structs com a tag `db`:

```go
type User struct {
	Name  string `db:"name"`       // Gera StringField
	Age   int    `db:"age"`        // Gera IntField
	Admin bool   `db:"is_admin"`   // Gera BoolField

	// Campos sem tag db são ignorados
	cachedValue string

	// Tag db:"-" ignora o campo explicitamente
	Password string `db:"-"`
}
```

## Tipos Suportados

| Tipo Go | Campo Gerado |
|---------|--------------|
| `string` | `query.StringField` |
| `int` | `query.IntField` |
| `int64` | `query.Int64Field` |
| `bool` | `query.BoolField` |
| `float64` | `query.Float64Field` |
| `core.Optional[string]` | `query.OptionalStringField` |
| `core.Optional[int]` | `query.OptionalIntField` |
| `core.Optional[int64]` | `query.OptionalInt64Field` |
| `core.Optional[bool]` | `query.OptionalBoolField` |
| `core.Optional[float64]` | `query.OptionalFloat64Field` |

## Integração com CI/CD

Adicione no seu workflow:

```bash
# Gerar código
go generate ./...

# Ou rode genus generate diretamente
genus generate ./models

# Verificar se há mudanças não comitadas
git diff --exit-code
```

## Comparação com Outras Ferramentas

### vs GORM
- **Genus**: Zero reflection em queries, campos gerados em compile-time
- **GORM**: Usa reflection em runtime para descobrir campos

### vs sqlboiler
- **Genus**: Gera campos tipados, mantém structs manuais
- **sqlboiler**: Gera structs completas a partir do schema do DB

### vs Squirrel
- **Genus**: Type-safe query builder com campos gerados
- **Squirrel**: Query builder baseado em strings

## Troubleshooting

### Arquivo não foi gerado
- Verifique se a struct tem campos com tag `db`
- Verifique se o arquivo não termina com `_test.go` ou `.gen.go`

### Erros de compilação após gerar
- Rode `go mod tidy` para garantir que as dependências estão corretas
- Verifique se os imports estão corretos no arquivo gerado

### Tipo não suportado
- Atualmente apenas tipos primitivos e `core.Optional[T]` são suportados
- Para tipos customizados, use `string` ou implemente `sql.Scanner`
