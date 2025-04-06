\c code_analyser

-- Migration to split the insights table into separate tables for each insight type
-- This will make querying and organizing data easier

-- Create function_insights table
CREATE TABLE IF NOT EXISTS code_analyzer.function_insights (
    id SERIAL PRIMARY KEY,
    repository_id BIGINT NOT NULL,
    function_id BIGINT NOT NULL,
    data JSONB NOT NULL, -- Store structured insight data
    model VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT fk_function_repository FOREIGN KEY (repository_id) REFERENCES code_analyzer.repositories(id) ON DELETE CASCADE,
    CONSTRAINT fk_function FOREIGN KEY (function_id) REFERENCES code_analyzer.repository_functions(id) ON DELETE CASCADE
);

-- Create symbol_insights table
CREATE TABLE IF NOT EXISTS code_analyzer.symbol_insights (
    id SERIAL PRIMARY KEY,
    repository_id BIGINT NOT NULL,
    symbol_id BIGINT NOT NULL,
    data JSONB NOT NULL, -- Store structured insight data
    model VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT fk_symbol_repository FOREIGN KEY (repository_id) REFERENCES code_analyzer.repositories(id) ON DELETE CASCADE,
    CONSTRAINT fk_symbol FOREIGN KEY (symbol_id) REFERENCES code_analyzer.repository_symbols(id) ON DELETE CASCADE
);

-- Create struct_insights table (for specialized struct analysis)
CREATE TABLE IF NOT EXISTS code_analyzer.struct_insights (
    id SERIAL PRIMARY KEY,
    repository_id BIGINT NOT NULL,
    symbol_id BIGINT NOT NULL, -- Structs are symbols in Go
    data JSONB NOT NULL, -- Store structured insight data
    model VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT fk_struct_repository FOREIGN KEY (repository_id) REFERENCES code_analyzer.repositories(id) ON DELETE CASCADE,
    CONSTRAINT fk_struct_symbol FOREIGN KEY (symbol_id) REFERENCES code_analyzer.repository_symbols(id) ON DELETE CASCADE
);

-- Create file_insights table
CREATE TABLE IF NOT EXISTS code_analyzer.file_insights (
    id SERIAL PRIMARY KEY,
    repository_id BIGINT NOT NULL,
    file_id BIGINT NOT NULL,
    data JSONB NOT NULL, -- Store structured insight data
    model VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT fk_file_repository FOREIGN KEY (repository_id) REFERENCES code_analyzer.repositories(id) ON DELETE CASCADE,
    CONSTRAINT fk_file FOREIGN KEY (file_id) REFERENCES code_analyzer.repository_files(id) ON DELETE CASCADE
);

-- Create repository_insights table
CREATE TABLE IF NOT EXISTS code_analyzer.repository_insights (
    id SERIAL PRIMARY KEY,
    repository_id BIGINT NOT NULL,
    data JSONB NOT NULL, -- Store structured insight data
    model VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT fk_repository FOREIGN KEY (repository_id) REFERENCES code_analyzer.repositories(id) ON DELETE CASCADE
);

-- Add indexes for better query performance
CREATE INDEX idx_function_insights_repository_id ON code_analyzer.function_insights(repository_id);
CREATE INDEX idx_function_insights_function_id ON code_analyzer.function_insights(function_id);

CREATE INDEX idx_symbol_insights_repository_id ON code_analyzer.symbol_insights(repository_id);
CREATE INDEX idx_symbol_insights_symbol_id ON code_analyzer.symbol_insights(symbol_id);

CREATE INDEX idx_struct_insights_repository_id ON code_analyzer.struct_insights(repository_id);
CREATE INDEX idx_struct_insights_symbol_id ON code_analyzer.struct_insights(symbol_id);

CREATE INDEX idx_file_insights_repository_id ON code_analyzer.file_insights(repository_id);
CREATE INDEX idx_file_insights_file_id ON code_analyzer.file_insights(file_id);

CREATE INDEX idx_repository_insights_repository_id ON code_analyzer.repository_insights(repository_id);

-- -- Migrate data from the old insights table to the new tables
-- -- Function insights migration
-- INSERT INTO code_analyzer.function_insights (repository_id, function_id, data, model, created_at, updated_at)
-- SELECT repository_id, function_id, data::jsonb, model, created_at, updated_at
-- FROM code_analyzer.insights
-- WHERE function_id IS NOT NULL AND type = 'function';

-- -- Symbol insights migration
-- INSERT INTO code_analyzer.symbol_insights (repository_id, symbol_id, data, model, created_at, updated_at)
-- SELECT repository_id, symbol_id, data::jsonb, model, created_at, updated_at
-- FROM code_analyzer.insights
-- WHERE symbol_id IS NOT NULL AND type = 'symbol';

-- -- Struct insights migration
-- INSERT INTO code_analyzer.struct_insights (repository_id, symbol_id, data, model, created_at, updated_at)
-- SELECT repository_id, symbol_id, data::jsonb, model, created_at, updated_at
-- FROM code_analyzer.insights
-- WHERE symbol_id IS NOT NULL AND type = 'struct';

-- -- File insights migration
-- INSERT INTO code_analyzer.file_insights (repository_id, file_id, data, model, created_at, updated_at)
-- SELECT repository_id, file_id, data::jsonb, model, created_at, updated_at
-- FROM code_analyzer.insights
-- WHERE file_id IS NOT NULL AND type = 'file';

-- -- Repository insights migration
-- INSERT INTO code_analyzer.repository_insights (repository_id, data, model, created_at, updated_at)
-- SELECT repository_id, data::jsonb, model, created_at, updated_at
-- FROM code_analyzer.insights
-- WHERE type = 'repository';

-- Add comments to the tables for better documentation
COMMENT ON TABLE code_analyzer.function_insights IS 'Stores LLM-generated insights for functions';
COMMENT ON TABLE code_analyzer.symbol_insights IS 'Stores LLM-generated insights for symbols';
COMMENT ON TABLE code_analyzer.struct_insights IS 'Stores LLM-generated insights for struct types';
COMMENT ON TABLE code_analyzer.file_insights IS 'Stores LLM-generated insights for files';
COMMENT ON TABLE code_analyzer.repository_insights IS 'Stores LLM-generated insights for repositories';

-- NOTE: After verifying the migration worked correctly, you can optionally 
-- drop the original insights table with:
-- DROP TABLE IF EXISTS code_analyzer.insights;

-- Down migration (if needed to rollback)
-- DROP TABLE IF EXISTS code_analyzer.function_insights;
-- DROP TABLE IF EXISTS code_analyzer.symbol_insights; 
-- DROP TABLE IF EXISTS code_analyzer.struct_insights;
-- DROP TABLE IF EXISTS code_analyzer.file_insights;
-- DROP TABLE IF EXISTS code_analyzer.repository_insights;
