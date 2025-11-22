# Migrations Example

Este exemplo demonstra o sistema completo de migrations do Genus, incluindo:

1. **AutoMigrate** - Criação rápida de tabelas para desenvolvimento
2. **Manual Migrations** - Controle total com versionamento para produção
3. **Rollback** - Reverter migrations
4. **CreateTableMigration** - Helper para gerar migrations a partir de structs

## Conceitos

### AutoMigrate vs Manual Migrations

| Feature | AutoMigrate | Manual Migrations |
|---------|-------------|-------------------|
| Uso | Desenvolvimento | Produção |
| Controle | Automático | Manual completo |
| Versionamento | Não | Sim |
| Rollback | Não | Sim |
| Customização | Limitada | Total |

### Quando Usar Cada Um

**AutoMigrate:**
- ✅ Protótipos rápidos
- ✅ Testes locais
- ✅ Desenvolvimento inicial
- ❌ Produção
- ❌ Equipes grandes

**Manual Migrations:**
- ✅ Produção
- ✅ Equipes grandes
- ✅ Controle de versão
- ✅ Rollback necessário
- ✅ Modificações complexas

## Executar o Exemplo

### Pré-requisitos

```bash
# Iniciar PostgreSQL
docker run -d \
  --name postgres-genus \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:latest

# Criar banco de dados
docker exec -it postgres-genus createdb -U postgres genus_migrations
```

### Executar

```bash
go run main.go
```

### Saída Esperada

```
=== Genus Migrations Example ===

1. AutoMigrate (Development)
   Creating tables from structs...
   ✅ Tables created successfully!

   Created user with ID: 1

2. Manual Migrations (Production)
   Setting up migrator...

   Migration Status (before):
   [ ] 1: create_users_table
   [ ] 2: add_users_indexes
   [ ] 3: create_products_table

   Applying migrations...
   [GENUS] Applying migration 1: create_users_table
   ✅ Applied migration 1: create_users_table
   [GENUS] Applying migration 2: add_users_indexes
   ✅ Applied migration 2: add_users_indexes
   [GENUS] Applying migration 3: create_products_table
   ✅ Applied migration 3: create_products_table

   Migration Status (after):
   [✓] 1: create_users_table
   [✓] 2: add_users_indexes
   [✓] 3: create_products_table

3. Rollback Demo
   Reverting last migration...
   [GENUS] Reverting migration 3: create_products_table
   ✅ Reverted migration 3: create_products_table

   Migration Status (after rollback):
   [✓] 1: create_users_table
   [✓] 2: add_users_indexes
   [ ] 3: create_products_table

4. CreateTableMigration Helper
   Creating migration from struct...
   ✅ Categories table created from struct!

=== Example completed successfully! ===
```

## 1. AutoMigrate

AutoMigrate cria tabelas automaticamente a partir de structs:

```go
import "github.com/gabrieldias/genus/migrate"

type User struct {
    core.Model
    Name  string `db:"name"`
    Email string `db:"email"`
}

// Criar tabela automaticamente
err := migrate.AutoMigrate(ctx, db, dialect, User{})
```

**O que acontece:**
- Analisa a struct User
- Extrai campos com tag `db`
- Gera CREATE TABLE apropriado
- Executa no banco de dados

**Limitações:**
- Não faz alterações em tabelas existentes
- Não versiona mudanças
- Não permite rollback

## 2. Manual Migrations

Migrations manuais dão controle total:

```go
import "github.com/gabrieldias/genus/migrate"

// Criar migrator
migrator := migrate.New(db, dialect, logger, migrate.Config{})

// Definir migration
migration := migrate.Migration{
    Version: 1,
    Name:    "create_users_table",
    Up: func(ctx context.Context, db *sql.DB, dialect core.Dialect) error {
        query := `
            CREATE TABLE users (
                id SERIAL PRIMARY KEY,
                name VARCHAR(255) NOT NULL,
                email VARCHAR(255) UNIQUE
            );
        `
        _, err := db.ExecContext(ctx, query)
        return err
    },
    Down: func(ctx context.Context, db *sql.DB, dialect core.Dialect) error {
        _, err := db.ExecContext(ctx, "DROP TABLE users")
        return err
    },
}

// Registrar e aplicar
migrator.Register(migration)
migrator.Up(ctx)
```

**Vantagens:**
- ✅ Versionamento automático
- ✅ Rollback com `Down()`
- ✅ Controle total do SQL
- ✅ Funciona com todos os bancos

## 3. Usando o CLI

### Criar Nova Migration

```bash
genus migrate create add_users_table
```

Gera arquivo `migrations/1234567890_add_users_table.go`:

```go
var Migration1234567890 = migrate.Migration{
    Version: 1234567890,
    Name:    "add_users_table",
    Up: func(ctx context.Context, db *sql.DB, dialect core.Dialect) error {
        // Implementar
        return nil
    },
    Down: func(ctx context.Context, db *sql.DB, dialect core.Dialect) error {
        // Implementar
        return nil
    },
}
```

### Aplicar Migrations

```bash
# Aplicar todas as pendentes
genus migrate up

# Reverter última
genus migrate down

# Ver status
genus migrate status
```

**Output do `genus migrate status`:**
```
Migration Status:
================
[✓] 1: create_users_table
[✓] 2: add_users_indexes
[ ] 3: create_products_table
```

## 4. CreateTableMigration Helper

Helper para gerar migrations a partir de structs:

```go
type Product struct {
    core.Model
    Name  string  `db:"name"`
    Price float64 `db:"price"`
}

// Criar migration automaticamente
migration := migrate.CreateTableMigration(
    1,
    "create_products_table",
    Product{},
)

migrator.Register(migration)
migrator.Up(ctx)
```

**Benefícios:**
- Combina conveniência de AutoMigrate com controle de migrations manuais
- Gera Up() e Down() automaticamente
- Permite versionamento e rollback

## Estrutura de Migrations

### Tabela de Controle

Genus cria uma tabela `schema_migrations` para tracking:

```sql
CREATE TABLE schema_migrations (
    version BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP NOT NULL
);
```

Cada migration aplicada é registrada:

| version | name | applied_at |
|---------|------|------------|
| 1 | create_users_table | 2024-01-15 10:30:00 |
| 2 | add_users_indexes | 2024-01-15 10:31:00 |

### Versionamento

Versões são números inteiros crescentes:

```go
migrations := []migrate.Migration{
    {Version: 1, Name: "create_users"},
    {Version: 2, Name: "add_email_index"},
    {Version: 3, Name: "create_products"},
}
```

**Boas práticas:**
- Use timestamps para versões únicas: `time.Now().Unix()`
- Ou sequencial simples: 1, 2, 3, ...
- Nunca edite migrations já aplicadas em produção

## Workflows Comuns

### Desenvolvimento Local

```go
// Rápido e fácil
migrate.AutoMigrate(ctx, db, dialect,
    User{},
    Product{},
    Order{},
)
```

### Deploy em Produção

```bash
# 1. Criar migration
genus migrate create add_new_feature

# 2. Implementar Up() e Down()
# (editar arquivo gerado)

# 3. Testar localmente
genus migrate status
genus migrate up

# 4. Commit ao git
git add migrations/
git commit -m "Add migration for new feature"

# 5. Deploy
# CI/CD roda: genus migrate up
```

### Rollback em Produção

```bash
# Ver o que está aplicado
genus migrate status

# Reverter última migration
genus migrate down

# Verificar
genus migrate status
```

## Boas Práticas

### 1. Sempre Implemente Down()

```go
// ❌ Ruim
Migration{
    Version: 1,
    Name: "create_users",
    Up: func(...) error { /* ... */ },
    Down: nil, // Sem rollback!
}

// ✅ Bom
Migration{
    Version: 1,
    Name: "create_users",
    Up: func(...) error { /* ... */ },
    Down: func(...) error { /* ... */ }, // Com rollback
}
```

### 2. Use Transações

Migrations já rodam em transações automaticamente. Se algo falhar, tudo é revertido.

### 3. Teste Rollback

```bash
# Aplicar
genus migrate up

# Testar rollback
genus migrate down

# Reaplicar
genus migrate up
```

### 4. Migrations Irreversíveis

Se uma migration não pode ser revertida (ex: deletar coluna com dados):

```go
Down: func(ctx context.Context, db *sql.DB, dialect core.Dialect) error {
    return fmt.Errorf("migration 3 cannot be reverted (data loss)")
}
```

### 5. Dados de Seed

Separe migrations de estrutura de dados de seed:

```go
// Migration - estrutura
Migration{
    Version: 1,
    Name: "create_users",
    Up: func(...) { /* CREATE TABLE */ },
}

// Seed - dados (separado)
func seedUsers(db *sql.DB) error {
    // INSERT dados iniciais
}
```

## Troubleshooting

### "migration already applied"

Migrations são idempotentes. Se já foi aplicada, `Up()` não faz nada.

### "migration table not found"

A primeira vez que rodar `migrate up`, a tabela é criada automaticamente.

### "failed to rollback"

Verifique se `Down()` está implementado corretamente.

### Diferenças entre dialects

```go
// Use o dialect parameter para queries portáveis
func(ctx context.Context, db *sql.DB, dialect core.Dialect) error {
    // ✅ Portável
    query := fmt.Sprintf("CREATE TABLE %s (...)",
        dialect.QuoteIdentifier("users"))

    // ❌ Hardcoded (apenas PostgreSQL)
    query := `CREATE TABLE "users" (...)`
}
```

## Próximos Passos

- Ver `examples/multi-database` para migrations cross-database
- Integrar com CI/CD
- Configurar `DATABASE_URL` via environment
- Criar migrations para seu schema
