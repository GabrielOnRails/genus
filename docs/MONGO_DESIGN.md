# Suporte a MongoDB: desenho e custo

Status: proposta, nada implementado.
Objetivo deste documento: decidir **se** e **como** antes de escrever código.

---

## 1. O ponto de partida errado

A leitura intuitiva é "Mongo é mais um dialeto, entra em `dialects/mongo/`". Isso
não funciona. A interface de dialeto atual é inteiramente SQL:

```go
// core/interfaces.go
type Dialect interface {
    Placeholder(n int) string          // $1, ?         -> Mongo não tem placeholder
    QuoteIdentifier(name string) string // "users"      -> Mongo não tem identificador citado
    GetType(goType string) string       // VARCHAR(255) -> Mongo não tem DDL
}

type Executor interface {                       // é a interface do database/sql
    ExecContext(ctx, query string, args ...interface{}) (sql.Result, error)
    QueryContext(ctx, query string, args ...interface{}) (*sql.Rows, error)
    QueryRowContext(ctx, query string, args ...interface{}) *sql.Row
}
```

Os três métodos de `Executor` recebem **uma string de query** e devolvem
`*sql.Rows`. Um backend Mongo não tem string de query nem rows: tem documento de
filtro BSON, cursor e pipeline de agregação. Implementar `Executor` para Mongo
significaria gerar SQL e traduzir SQL para BSON, o que é a pior das opções
(parser de SQL dentro do ORM).

Escala do acoplamento hoje:

- 29 arquivos não-teste montam SQL como string via `strings.Builder`
- 16 pacotes importam `database/sql`
- `core.getID` (core/db.go:467) retorna `int64` e todo `Update`/`Delete` monta
  `WHERE id = <placeholder>`; Mongo usa `ObjectID` (12 bytes) como `_id`

Conclusão: Mongo é um **segundo backend atrás de uma abstração que ainda não
existe**, não um pacote em `dialects/`.

---

## 2. O que faz esse trabalho valer a pena

O driver oficial do Mongo em Go resolve tipo em runtime. Marshalling de BSON
para struct falha na execução, não na compilação, e é a reclamação mais comum de
quem usa o driver nativo em projeto grande.

Isso é exatamente o que o Genus já faz bem no SQL: `UserFields.Name.Eq("x")`
erra na compilação. Aplicar a mesma ideia a filtro BSON é um diferencial real,
porque hoje não existe no ecossistema Go:

```go
// driver nativo: chave é string, valor é interface{}, erro aparece em runtime
filter := bson.M{"emial": bson.M{"$gt": 30}}   // typo e tipo errado, compila

// alvo: chave e tipo validados pelo compilador
users, err := genus.Collection[User](db).
    Where(UserFields.Age.Gt(30)).
    Find(ctx)
```

O ganho não é performance (o driver nativo já é rápido). É **eliminar erro de
tipo em runtime**. Esse é o pitch, e ele é honesto.

---

## 3. Arquitetura proposta: representação intermediária

O builder para de emitir SQL diretamente e passa a emitir uma estrutura neutra,
que cada backend compila para o seu formato.

```
Builder[T]  ──emite──>  QuerySpec  ──compila──>  SQLCompiler   -> string + args
                        (neutro)                  MongoCompiler -> bson.D + options
```

```go
// Neutro: sem string de SQL, sem BSON
type QuerySpec struct {
    Source     string        // tabela ou collection
    Projection []string
    Filter     []Predicate   // o que hoje é Condition/ConditionGroup
    Sort       []SortKey
    Limit      *int
    Offset     *int
    Joins      []JoinSpec    // SQL apenas; MongoCompiler rejeita
}

type Backend interface {
    Find(ctx context.Context, spec QuerySpec, dest any) error
    Update(ctx context.Context, spec QuerySpec, assignments []Assignment) (int64, error)
    Delete(ctx context.Context, spec QuerySpec) (int64, error)
    Capabilities() Capabilities
}

// O que cada backend suporta, verificado em runtime com erro claro
type Capabilities struct {
    Joins        bool
    Transactions bool
    Migrations   bool
    RawSQL       bool
}
```

Ponto central do desenho: **`Capabilities` explícito**. Sem isso, o Genus vira
uma abstração que promete tudo e falha silenciosamente em metade. Chamar `Join`
num backend Mongo tem que devolver erro nomeado, não SQL inválido.

Os fields tipados (`query/field.go`) e os `Assignment` (`query/assignment.go`)
**já são neutros**: produzem `Condition{Field, Operator, Value}` e
`Assignment{Field, Value}`, sem nada de SQL. Essa parte não muda, e é o que
sustenta o valor descrito na seção 2.

---

## 4. O que não dá para suportar

Honestidade aqui evita retrabalho e README mentiroso.

| Recurso | Situação em Mongo |
|---|---|
| `Join`, `LeftJoin`, `RightJoin` | Não existe. `$lookup` só cobre um subconjunto e tem custo alto. Fica fora. |
| `migrate/` (AutoMigrate, versionado) | Sem schema, perde sentido. No máximo gerência de índice. |
| `ObjectID` vs `int64` | `core.getID` retorna `int64`. Precisa virar um tipo de identidade abstrato. |
| Transações | Só com replica set. Vai para `Capabilities`. |
| `Preload` | Precisa ser reescrito sobre `$lookup` ou query extra. |
| `sharding/` | Mongo já faz sharding nativo. O pacote perde propósito. |
| `GenusUltra` (scan sem reflection) | O ganho vem de `rows.Scan`. Em BSON o gargalo é o unmarshal, otimização diferente. |
| "Zero dependencies" | Morre. O driver do Mongo é dependência pesada. Ver seção 6. |

---

## 5. Impacto nos pacotes atuais

| Pacote | Impacto |
|---|---|
| `query/field.go`, `query/assignment.go` | Nenhum. Já são neutros. |
| `query/builder.go` | Reescrever: emitir `QuerySpec` em vez de string. Alto. |
| `query/mutation.go` | Idem. A guarda de `WHERE` vazio vale igual em Mongo. |
| `core/db.go` | `Create`/`Update`/`Delete` assumem `id int64` e SQL. Alto. |
| `core/interfaces.go` | `Executor` e `Dialect` deixam de ser o contrato central. |
| `migrate/`, `sharding/` | Ficam restritos a backend SQL. |
| `dialects/*` | Passam a ser compiladores de `QuerySpec`, não dialetos. |

Isso é uma mudança de major version, com risco real de regressão nos 3 dialetos
que hoje funcionam.

---

## 6. Pré-requisito: separar dependências em submódulos

Hoje `go.mod` da raiz já traz `gorm.io/gorm` (por causa de `benchmarks/`) e 5
pacotes de OpenTelemetry como dependências **diretas**, apesar de o README
anunciar "Zero dependencies". Somar o driver do Mongo a isso agrava o problema.

Antes de qualquer código de Mongo:

- `benchmarks/go.mod` separado, para GORM sair do módulo raiz
- `tracing/` e cada driver em submódulo próprio
- backend Mongo como submódulo `genus/mongo`, para quem usa só SQL não baixar o driver

Sem isso, adicionar Mongo torna o claim do README indefensável.

---

## 7. Fases sugeridas

| Fase | Entrega | Risco |
|---|---|---|
| 0 | Submódulos (seção 6) e correção do claim do README | Baixo |
| 1 | `QuerySpec` + `SQLCompiler`, mantendo a API pública e os testes atuais verdes | Médio |
| 2 | `Capabilities` e erro nomeado para recurso não suportado | Baixo |
| 3 | `MongoBackend`: `Find`, `Where`, `Update`, `Delete`, `Count` | Médio |
| 4 | Identidade abstrata (`ObjectID`), índices, `Preload` via `$lookup` | Alto |

A fase 1 é a que decide o projeto: se `QuerySpec` conseguir substituir a
montagem de SQL sem quebrar os testes existentes, o resto é incremental. Se não
conseguir, Mongo não sai sem reescrita completa.

---

## 8. Decisões abertas

1. **Posicionamento.** "ORM SQL type-safe" ou "data mapper poliglota"? Muda o
   README, o pitch e as interfaces centrais.
2. **Compatibilidade.** Manter a API pública atual na fase 1 ou aceitar quebra?
3. **Escopo do Mongo.** CRUD com filtro tipado resolve o problema real (tipo em
   runtime). Vale perseguir agregação e `$lookup`, ou declarar fora de escopo?
4. **Ordem.** Antes de Mongo existem dois furos no caminho SQL: ausência de
   teste de integração contra Postgres e MySQL reais, e ausência de benchmark de
   join e preload. Mongo aumenta a superfície de uma base que ainda não é
   verificada nos backends que já suporta.
