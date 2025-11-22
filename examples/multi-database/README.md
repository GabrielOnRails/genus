# Multi-Database Example

Este exemplo demonstra como usar o Genus com diferentes bancos de dados (PostgreSQL, MySQL e SQLite) usando o sistema de dialects.

## Bancos Suportados

Genus 1.x suporta três bancos de dados através de dialects:

1. **PostgreSQL** - `dialects/postgres`
2. **MySQL** - `dialects/mysql`
3. **SQLite** - `dialects/sqlite`

## Como Funciona

Cada banco de dados tem suas próprias convenções:

### PostgreSQL
- **Placeholders:** `$1`, `$2`, `$3`, etc.
- **Identificadores:** Aspas duplas `"table_name"`
- **Auto-increment:** `SERIAL`

### MySQL
- **Placeholders:** `?` (posicional)
- **Identificadores:** Backticks `` `table_name` ``
- **Auto-increment:** `AUTO_INCREMENT`

### SQLite
- **Placeholders:** `?` (posicional)
- **Identificadores:** Aspas duplas `"table_name"`
- **Auto-increment:** `AUTOINCREMENT`
- **Nota:** Não tem tipo `BOOLEAN` nativo (usa `INTEGER`)

## Usando Diferentes Dialects

```go
import (
    "database/sql"
    "github.com/GabrielOnRails/genus"
    "github.com/GabrielOnRails/genus/dialects/postgres"
    "github.com/GabrielOnRails/genus/dialects/mysql"
    "github.com/GabrielOnRails/genus/dialects/sqlite"
)

// PostgreSQL
db, _ := sql.Open("postgres", "...")
g := genus.New(db, postgres.New(), logger)

// MySQL
db, _ := sql.Open("mysql", "...")
g := genus.New(db, mysql.New(), logger)

// SQLite
db, _ := sql.Open("sqlite3", ":memory:")
g := genus.New(db, sqlite.New(), logger)
```

## Instalação de Drivers

Você precisa instalar os drivers SQL apropriados:

```bash
# PostgreSQL
go get github.com/lib/pq

# MySQL
go get github.com/go-sql-driver/mysql

# SQLite
go get github.com/mattn/go-sqlite3
```

## Executar o Exemplo

### Pré-requisitos

#### PostgreSQL
```bash
# Docker
docker run --name postgres-genus \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=genus_test \
  -p 5432:5432 \
  -d postgres:latest

# Ou instale localmente e crie o banco:
createdb genus_test
```

#### MySQL
```bash
# Docker
docker run --name mysql-genus \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_DATABASE=genus_test \
  -p 3306:3306 \
  -d mysql:latest

# Ou instale localmente e crie o banco:
mysql -u root -p -e "CREATE DATABASE genus_test"
```

#### SQLite
Não requer setup - usa banco em memória (`:memory:`)

### Executar

```bash
# Instalar dependências
go mod download

# Executar
go run main.go
```

### Saída Esperada

```
=== Genus Multi-Database Example ===

1. PostgreSQL
   Testing CRUD operations on PostgreSQL...
   - Created user with ID: 1
   - Found 1 user(s)
   - User: Alice (email: alice@example.com, age: 25)
   - Updated user age to 26
   - Active users count: 1
   - Deleted user
   ✅ PostgreSQL working!

2. MySQL
   Testing CRUD operations on MySQL...
   - Created user with ID: 1
   - Found 1 user(s)
   - User: Alice (email: alice@example.com, age: 25)
   - Updated user age to 26
   - Active users count: 1
   - Deleted user
   ✅ MySQL working!

3. SQLite
   Testing CRUD operations on SQLite...
   - Created user with ID: 1
   - Found 1 user(s)
   - User: Alice (email: alice@example.com, age: 25)
   - Updated user age to 26
   - Active users count: 1
   - Deleted user
   ✅ SQLite working!

=== Example completed! ===
```

## Queries Geradas

O Genus gera queries específicas para cada banco:

### PostgreSQL
```sql
SELECT * FROM "users" WHERE name = $1
INSERT INTO "users" (name, email, age) VALUES ($1, $2, $3)
UPDATE "users" SET name = $1, age = $2 WHERE id = $3
DELETE FROM "users" WHERE id = $1
```

### MySQL
```sql
SELECT * FROM `users` WHERE name = ?
INSERT INTO `users` (name, email, age) VALUES (?, ?, ?)
UPDATE `users` SET name = ?, age = ? WHERE id = ?
DELETE FROM `users` WHERE id = ?
```

### SQLite
```sql
SELECT * FROM "users" WHERE name = ?
INSERT INTO "users" (name, email, age) VALUES (?, ?, ?)
UPDATE "users" SET name = ?, age = ? WHERE id = ?
DELETE FROM "users" WHERE id = ?
```

## Abstraindo o Banco de Dados

O código do usuário permanece o mesmo independente do banco:

```go
// Este código funciona com PostgreSQL, MySQL e SQLite!
users, err := g.Table[User]().
    Where(UserFields.Name.Eq("Alice")).
    Where(UserFields.Age.Gt(18)).
    OrderByDesc("created_at").
    Limit(10).
    Find(ctx)
```

O dialect se encarrega de:
- Gerar placeholders corretos (`$1` vs `?`)
- Escapar identificadores (`` ` `` vs `"`)
- Mapear tipos Go para SQL apropriados

## Migrando Entre Bancos

Para migrar de um banco para outro:

1. Atualize a connection string
2. Troque o dialect
3. Ajuste o schema (tipos de dados, auto-increment)
4. **Nenhuma mudança no código do ORM!**

```go
// De PostgreSQL
g := genus.New(db, postgres.New(), logger)

// Para MySQL
g := genus.New(db, mysql.New(), logger)

// Queries continuam funcionando!
```

## Limitações por Banco

### PostgreSQL
- ✅ Suporte completo a todos os recursos
- ✅ Tipos avançados (JSON, UUID, Arrays)
- ✅ Transações robustas

### MySQL
- ✅ Suporte completo aos recursos básicos
- ⚠️ Diferenças em `TIMESTAMP` vs `DATETIME`
- ⚠️ Menos tipos avançados que PostgreSQL

### SQLite
- ✅ Perfeito para desenvolvimento e testes
- ✅ Zero configuração
- ⚠️ Sistema de tipos mais simples
- ⚠️ `BOOLEAN` → `INTEGER` (0/1)
- ⚠️ Menos recursos de concorrência

## Boas Práticas

1. **Use variáveis de ambiente** para connection strings
   ```go
   dsn := os.Getenv("DATABASE_URL")
   ```

2. **Configure timeouts apropriados**
   ```go
   db.SetMaxOpenConns(25)
   db.SetMaxIdleConns(5)
   db.SetConnMaxLifetime(5 * time.Minute)
   ```

3. **Use context para timeouts**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   users, err := g.Table[User]().Find(ctx)
   ```

4. **Teste com SQLite, produza com PostgreSQL/MySQL**
   ```go
   var dialect core.Dialect
   if isTest {
       dialect = sqlite.New()
   } else {
       dialect = postgres.New()
   }
   ```

## Troubleshooting

### PostgreSQL: "connection refused"
- Verifique se o servidor está rodando: `pg_isready`
- Verifique as credenciais e hostname

### MySQL: "connection refused"
- Verifique se o MySQL está rodando: `mysqladmin ping`
- Verifique se o usuário tem permissões

### SQLite: "unable to open database file"
- Verifique permissões de escrita no diretório
- Use `:memory:` para testes rápidos

### Todos: "driver not found"
- Certifique-se de importar o driver:
  ```go
  import _ "github.com/lib/pq"
  import _ "github.com/go-sql-driver/mysql"
  import _ "github.com/mattn/go-sqlite3"
  ```
