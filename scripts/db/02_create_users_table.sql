-- Connect to the database
\c code_analyser

CREATE SCHEMA IF NOT EXISTS users;

-- Grant all privileges on users schema to code_analyser_user
GRANT ALL PRIVILEGES ON SCHEMA users TO code_analyser_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA users TO code_analyser_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA users TO code_analyser_user;
GRANT USAGE ON SCHEMA users TO code_analyser_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA users GRANT ALL PRIVILEGES ON TABLES TO code_analyser_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA users GRANT ALL PRIVILEGES ON SEQUENCES TO code_analyser_user;


-- Create users table
CREATE TABLE IF NOT EXISTS users.users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    active BOOLEAN DEFAULT TRUE,
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users.users(email);

-- Create index on role for role-based queries
CREATE INDEX IF NOT EXISTS idx_users_role ON users.users(role);

-- Create index on active status
CREATE INDEX IF NOT EXISTS idx_users_active ON users.users(active);
