-- Connect to the database
\c code_analyser

-- Create tables for Go code analyzer
CREATE SCHEMA IF NOT EXISTS code_analyzer;

-- Grant all privileges on code_analyzer schema to code_analyser_user
GRANT ALL PRIVILEGES ON SCHEMA code_analyzer TO code_analyser_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA code_analyzer TO code_analyser_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA code_analyzer TO code_analyser_user;
GRANT USAGE ON SCHEMA code_analyzer TO code_analyser_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA code_analyzer GRANT ALL PRIVILEGES ON TABLES TO code_analyser_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA code_analyzer GRANT ALL PRIVILEGES ON SEQUENCES TO code_analyser_user;

-- Table to store repositories
CREATE TABLE IF NOT EXISTS code_analyzer.repositories (
    id SERIAL PRIMARY KEY,
    kind VARCHAR(50) NOT NULL, -- "github", "gitlab", etc.
    url VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    owner VARCHAR(100) NOT NULL,
    local_path VARCHAR(255) NOT NULL,
    last_indexed TIMESTAMP,
    index_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- "pending", "in_progress", "completed", "failed"
    index_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Table to store repository files
CREATE TABLE IF NOT EXISTS code_analyzer.repository_files (
    id SERIAL PRIMARY KEY,
    repository_id INTEGER NOT NULL REFERENCES code_analyzer.repositories(id) ON DELETE CASCADE,
    file_path VARCHAR(500) NOT NULL, -- Relative path within repo
    package VARCHAR(100) NOT NULL, -- Go package name
    last_analyzed TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(repository_id, file_path)
);

-- Table to store repository functions
CREATE TABLE IF NOT EXISTS code_analyzer.repository_functions (
    id SERIAL PRIMARY KEY,
    repository_id INTEGER NOT NULL REFERENCES code_analyzer.repositories(id) ON DELETE CASCADE,
    file_id INTEGER NOT NULL REFERENCES code_analyzer.repository_files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    kind VARCHAR(50) NOT NULL, -- "function" or "method"
    receiver VARCHAR(255), -- For methods
    exported BOOLEAN NOT NULL,
    parameters JSONB, -- JSON array of parameters
    results JSONB, -- JSON array of results
    code_block TEXT, -- Full code
    line INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(repository_id, file_id, name, line)
);

-- Table to store function calls (normalized)
CREATE TABLE IF NOT EXISTS code_analyzer.function_calls (
    id SERIAL PRIMARY KEY,
    caller_id INTEGER NOT NULL REFERENCES code_analyzer.repository_functions(id) ON DELETE CASCADE,
    callee_name VARCHAR(255) NOT NULL, -- Function being called
    callee_package VARCHAR(255), -- Package of the function being called
    callee_id INTEGER REFERENCES code_analyzer.repository_functions(id) ON DELETE CASCADE, -- May be NULL for external calls
    line INTEGER NOT NULL, -- Line number where the call occurs
    parameters JSONB, -- Parameters passed to the function
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(caller_id, callee_name, line)
);

-- Table to store function references (normalized)
CREATE TABLE IF NOT EXISTS code_analyzer.function_references (
    id SERIAL PRIMARY KEY,
    function_id INTEGER NOT NULL REFERENCES code_analyzer.repository_functions(id) ON DELETE CASCADE,
    reference_type VARCHAR(50) NOT NULL, -- "declaration", "usage", "modification"
    file_id INTEGER NOT NULL REFERENCES code_analyzer.repository_files(id) ON DELETE CASCADE,
    line INTEGER NOT NULL,
    column_position INTEGER NOT NULL,
    context TEXT, -- Small code snippet for context
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(function_id, file_id, line, column_position)
);

-- Table to store statement analysis (normalized)
CREATE TABLE IF NOT EXISTS code_analyzer.function_statements (
    id SERIAL PRIMARY KEY,
    function_id INTEGER NOT NULL REFERENCES code_analyzer.repository_functions(id) ON DELETE CASCADE,
    statement_type VARCHAR(50) NOT NULL, -- "if", "for", "switch", "return", "assignment", etc.
    text TEXT NOT NULL, -- Text representation
    line INTEGER NOT NULL,
    conditions JSONB, -- Conditions (for if/loop/switch)
    variables JSONB, -- Variables used or defined
    calls JSONB, -- Function calls within this statement
    parent_statement_id INTEGER DEFAULT NULL REFERENCES code_analyzer.function_statements(id) ON DELETE CASCADE, -- For nested statements
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Table to store other repository symbols (variables, constants, types, etc.)
CREATE TABLE IF NOT EXISTS code_analyzer.repository_symbols (
    id SERIAL PRIMARY KEY,
    repository_id INTEGER NOT NULL REFERENCES code_analyzer.repositories(id) ON DELETE CASCADE,
    file_id INTEGER NOT NULL REFERENCES code_analyzer.repository_files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    kind VARCHAR(50) NOT NULL, -- "variable", "constant", "type", "struct", "interface"
    type VARCHAR(255),
    value TEXT,
    exported BOOLEAN NOT NULL,
    fields JSONB, -- JSON array of fields (for structs)
    methods JSONB, -- JSON array of methods
    line INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(repository_id, file_id, name, line)
);

-- Table to store symbol references (normalized)
CREATE TABLE IF NOT EXISTS code_analyzer.symbol_references (
    id SERIAL PRIMARY KEY,
    symbol_id INTEGER NOT NULL REFERENCES code_analyzer.repository_symbols(id) ON DELETE CASCADE,
    reference_type VARCHAR(50) NOT NULL, -- "declaration", "usage", "modification"
    file_id INTEGER NOT NULL REFERENCES code_analyzer.repository_files(id) ON DELETE CASCADE,
    line INTEGER NOT NULL,
    column_position INTEGER NOT NULL,
    context TEXT, -- Small code snippet for context
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(symbol_id, file_id, line, column_position)
);

-- Table to store file-level dependencies
CREATE TABLE IF NOT EXISTS code_analyzer.file_dependencies (
    id SERIAL PRIMARY KEY,
    repository_id INTEGER NOT NULL REFERENCES code_analyzer.repositories(id) ON DELETE CASCADE,
    file_id INTEGER NOT NULL REFERENCES code_analyzer.repository_files(id) ON DELETE CASCADE, 
    import_path VARCHAR(500) NOT NULL, -- The import path (e.g., "github.com/pkg/errors")
    alias VARCHAR(100), -- The alias used in the import (e.g., "errors")
    is_stdlib BOOLEAN NOT NULL DEFAULT false, -- Whether this is a standard library import
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(file_id, import_path)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_repository_files_repository_id ON code_analyzer.repository_files(repository_id);
CREATE INDEX IF NOT EXISTS idx_repository_functions_repository_id ON code_analyzer.repository_functions(repository_id);
CREATE INDEX IF NOT EXISTS idx_repository_functions_file_id ON code_analyzer.repository_functions(file_id);
CREATE INDEX IF NOT EXISTS idx_repository_symbols_repository_id ON code_analyzer.repository_symbols(repository_id);
CREATE INDEX IF NOT EXISTS idx_repository_symbols_file_id ON code_analyzer.repository_symbols(file_id);
CREATE INDEX IF NOT EXISTS idx_repository_functions_name ON code_analyzer.repository_functions(name);
CREATE INDEX IF NOT EXISTS idx_repository_symbols_name ON code_analyzer.repository_symbols(name);
CREATE INDEX IF NOT EXISTS idx_function_calls_caller_id ON code_analyzer.function_calls(caller_id);
CREATE INDEX IF NOT EXISTS idx_function_calls_callee_id ON code_analyzer.function_calls(callee_id);
CREATE INDEX IF NOT EXISTS idx_function_references_function_id ON code_analyzer.function_references(function_id);
CREATE INDEX IF NOT EXISTS idx_function_statements_function_id ON code_analyzer.function_statements(function_id);
CREATE INDEX IF NOT EXISTS idx_function_statements_parent_id ON code_analyzer.function_statements(parent_statement_id);
CREATE INDEX IF NOT EXISTS idx_symbol_references_symbol_id ON code_analyzer.symbol_references(symbol_id);
CREATE INDEX IF NOT EXISTS idx_file_dependencies_file_id ON code_analyzer.file_dependencies(file_id);
CREATE INDEX IF NOT EXISTS idx_file_dependencies_repository_id ON code_analyzer.file_dependencies(repository_id);
CREATE INDEX IF NOT EXISTS idx_file_dependencies_import_path ON code_analyzer.file_dependencies(import_path);
