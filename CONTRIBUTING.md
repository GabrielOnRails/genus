# Contribuindo para o Genus

Obrigado por considerar contribuir para o Genus! Este documento fornece diretrizes para contribuições.

## Como Contribuir

### 1. Reportar Bugs

Se você encontrar um bug, por favor abra uma issue com:

- Descrição clara do problema
- Passos para reproduzir
- Comportamento esperado vs atual
- Versão do Go e do Genus
- Sistema operacional

### 2. Sugerir Features

Antes de sugerir uma nova feature:

1. Verifique se já não existe uma issue sobre isso
2. Descreva claramente o caso de uso
3. Explique como se encaixa na filosofia do Genus (type-safety, simplicidade, etc.)

### 3. Pull Requests

#### Preparação

1. Fork o repositório
2. Crie uma branch para sua feature: `git checkout -b feature/minha-feature`
3. Faça suas mudanças
4. Adicione testes (se aplicável)
5. Certifique-se de que tudo compila: `go build ./...`
6. Commit suas mudanças: `git commit -m "feat: adiciona X"`

#### Diretrizes de Código

- **Simplicidade**: Código simples é melhor que código "inteligente"
- **Type-Safety**: Sempre priorize type-safety usando generics
- **Zero Magic**: Evite reflection sempre que possível
- **Context-Aware**: Funções públicas devem aceitar `context.Context`
- **Documentação**: Comente código público seguindo Go doc conventions
- **Testes**: Adicione testes para novas funcionalidades

#### Convenções de Commit

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - Nova feature
- `fix:` - Correção de bug
- `docs:` - Mudanças na documentação
- `refactor:` - Refatoração de código
- `test:` - Adicionar ou atualizar testes
- `chore:` - Tarefas de manutenção

Exemplos:
```
feat: adiciona suporte a MySQL
fix: corrige bug no scanner de structs aninhadas
docs: atualiza README com exemplos de transações
refactor: simplifica buildWhereClause
test: adiciona testes para IntField operators
```

### 4. Estrutura do Projeto

```
genus/
├── core/           # Core do ORM (DB, Model, interfaces)
├── query/          # Query builder e campos tipados
├── dialects/       # Dialetos de banco de dados
│   └── postgres/
├── examples/       # Exemplos de uso
└── docs/           # Documentação adicional (futura)
```

### 5. Diretrizes de Design

#### Use Generics Extensivamente

```go
// ✅ Bom - type-safe
func Find[T any](ctx context.Context) ([]T, error)

// ❌ Ruim - não type-safe
func Find(ctx context.Context, dest interface{}) error
```

#### Minimize Reflection

```go
// ✅ Bom - usa generics
type Builder[T any] struct {
    // ...
}

// ❌ Ruim - usa reflection excessivamente
func (b *Builder) Build(model interface{}) {
    typ := reflect.TypeOf(model)
    // ... muita reflection ...
}
```

#### Priorize Composição

```go
// ✅ Bom - composição
type User struct {
    core.Model  // Embedded
    Name string
}

// ❌ Ruim - herança (Go não tem, mas evite simular)
```

#### Context em Primeiro Lugar

```go
// ✅ Bom
func (b *Builder[T]) Find(ctx context.Context) ([]T, error)

// ❌ Ruim
func (b *Builder[T]) Find() ([]T, error)
```

### 6. Testes

#### Estrutura de Testes

- Testes unitários: `*_test.go` no mesmo pacote
- Testes de integração: `examples/testing/`
- Benchmarks: `Benchmark*` functions

#### Executar Testes

```bash
# Todos os testes
go test ./...

# Com coverage
go test -cover ./...

# Benchmarks
go test -bench=. ./...
```

### 7. Roadmap

Áreas que precisam de contribuições:

#### Alta Prioridade

- [ ] Dialetos: MySQL, SQLite
- [ ] Code generation para campos tipados
- [ ] Migrations
- [ ] Relações (HasMany, BelongsTo, ManyToMany)

#### Média Prioridade

- [ ] Hooks avançados (AfterCreate, BeforeUpdate, etc.)
- [ ] Soft deletes
- [ ] Query logging e debugging
- [ ] Prepared statements optimization

#### Baixa Prioridade

- [ ] Connection pooling configuration
- [ ] Query caching
- [ ] Metrics e observability

### 8. Processo de Review

1. Mantenedor(es) revisarão o PR
2. Feedback será dado construtivamente
3. Mudanças podem ser solicitadas
4. Após aprovação, será feito merge

### 9. Código de Conduta

- Seja respeitoso e construtivo
- Aceite críticas construtivas
- Foque no código, não nas pessoas
- Ajude outros contribuidores

### 10. Perguntas?

Se tiver dúvidas:

1. Verifique a [documentação](./README.md)
2. Leia a [arquitetura](./ARCHITECTURE.md)
3. Abra uma issue com a tag `question`

## Primeiros Passos para Contribuidores

### Setup do Ambiente

```bash
# 1. Clone o fork
git clone https://github.com/SEU-USUARIO/genus.git
cd genus

# 2. Adicione o upstream
git remote add upstream https://github.com/gabrieldias/genus.git

# 3. Crie uma branch
git checkout -b feature/minha-feature

# 4. Instale dependências
go mod download

# 5. Verifique que compila
go build ./...

# 6. Faça suas mudanças...

# 7. Commit e push
git add .
git commit -m "feat: minha feature"
git push origin feature/minha-feature

# 8. Abra um PR no GitHub
```

### Boas Primeiras Issues

Procure por issues com as tags:

- `good first issue` - Bom para iniciantes
- `help wanted` - Precisa de ajuda
- `documentation` - Melhorias na documentação

## Agradecimentos

Obrigado por contribuir para tornar o Genus melhor! 🚀
