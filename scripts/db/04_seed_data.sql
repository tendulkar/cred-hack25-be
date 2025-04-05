-- Connect to the database
\c code_analyser

-- Seed admin user (password is 'admin123' hashed with bcrypt)
INSERT INTO users.users (id, email, password, first_name, last_name, active, role, created_at, updated_at)
VALUES (
    uuid_generate_v4(),
    'admin@example.com',
    '$2a$10$3Qm7G.eV3SYcV8K1YbYmEOBQJYNrBxZN.f0FTiCZkRD9XLNHvZ5Uu',
    'Admin',
    'User',
    TRUE,
    'admin',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
) ON CONFLICT (email) DO NOTHING;

-- Seed regular user (password is 'user123' hashed with bcrypt)
INSERT INTO users.users (id, email, password, first_name, last_name, active, role, created_at, updated_at)
VALUES (
    uuid_generate_v4(),
    'user@example.com',
    '$2a$10$qH.qP.fI3ZN7Ebnk5a/JIeQh5VR9OHZ2Y5KhOXwE1M5jKpvFx.1Wy',
    'Regular',
    'User',
    TRUE,
    'user',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
) ON CONFLICT (email) DO NOTHING;

-- Seed a sample repository
INSERT INTO code_analysis.repositories (id, repo_url, owner, name, created_at, updated_at, analyzed_at)
VALUES (
    uuid_generate_v4(),
    'https://github.com/golang/go',
    'golang',
    'go',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    NULL
) ON CONFLICT (repo_url) DO NOTHING;
