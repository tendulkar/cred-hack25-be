-- Create the insights table for storing LLM-generated repository insights

-- Create the table
CREATE TABLE IF NOT EXISTS code_analyzer.insights (
    id SERIAL PRIMARY KEY,
    repository_id BIGINT NOT NULL,
    file_id BIGINT,
    function_id BIGINT,
    symbol_id BIGINT,
    path VARCHAR(1024),
    type VARCHAR(50) NOT NULL,
    data TEXT NOT NULL,
    model VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT fk_repository FOREIGN KEY (repository_id) REFERENCES code_analyzer.repositories(id) ON DELETE CASCADE,
    CONSTRAINT fk_file FOREIGN KEY (file_id) REFERENCES code_analyzer.repository_files(id) ON DELETE CASCADE,
    CONSTRAINT fk_function FOREIGN KEY (function_id) REFERENCES code_analyzer.repository_functions(id) ON DELETE CASCADE,
    CONSTRAINT fk_symbol FOREIGN KEY (symbol_id) REFERENCES code_analyzer.repository_symbols(id) ON DELETE CASCADE
);

-- Indexes for faster queries
CREATE INDEX idx_insights_repository_id ON code_analyzer.insights(repository_id);
CREATE INDEX idx_insights_file_id ON code_analyzer.insights(file_id);
CREATE INDEX idx_insights_function_id ON code_analyzer.insights(function_id);
CREATE INDEX idx_insights_symbol_id ON code_analyzer.insights(symbol_id);
CREATE INDEX idx_insights_type ON code_analyzer.insights(type);

-- Add a comment to the table
COMMENT ON TABLE code_analyzer.insights IS 'Stores LLM-generated insights for repository components';

-- Down migration
-- DROP TABLE IF EXISTS code_analyzer.insights;
