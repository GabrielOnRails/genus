-- Schema SQL para testar o exemplo básico do Genus

-- Cria a tabela users
CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    age        INTEGER NOT NULL,
    is_active  BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insere dados de exemplo
INSERT INTO users (name, email, age, is_active) VALUES
    ('Alice', 'alice@example.com', 28, true),
    ('Bob', 'bob@example.com', 32, true),
    ('Charlie', 'charlie@test.com', 24, false),
    ('Diana', 'diana@example.com', 35, true),
    ('Eve', 'eve@example.com', 22, true),
    ('Frank', 'frank@test.com', 40, false),
    ('Grace', 'grace@example.com', 30, true);

-- Cria índices
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_age ON users(age);
CREATE INDEX idx_users_is_active ON users(is_active);
