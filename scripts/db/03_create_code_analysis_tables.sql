-- Connect to the database
\c code_analyser

CREATE SCHEMA IF NOT EXISTS code_analysis;

-- Grant all privileges on code_analysis schema to code_analyser_user
GRANT ALL PRIVILEGES ON SCHEMA code_analysis TO code_analyser_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA code_analysis TO code_analyser_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA code_analysis TO code_analyser_user;
GRANT USAGE ON SCHEMA code_analysis TO code_analyser_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA code_analysis GRANT ALL PRIVILEGES ON TABLES TO code_analyser_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA code_analysis GRANT ALL PRIVILEGES ON SEQUENCES TO code_analyser_user;

-- Create repositories table
CREATE TABLE IF NOT EXISTS code_analysis.repositories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_url VARCHAR(255) UNIQUE NOT NULL,
    owner VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    analyzed_at TIMESTAMP WITH TIME ZONE
);

-- Create files table
CREATE TABLE IF NOT EXISTS code_analysis.files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repository_id UUID NOT NULL REFERENCES code_analysis.repositories(id) ON DELETE CASCADE,
    path VARCHAR(1000) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(repository_id, path)
);

-- Create dependencies table
CREATE TABLE IF NOT EXISTS code_analysis.dependencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES code_analysis.files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create global_vars table
CREATE TABLE IF NOT EXISTS code_analysis.global_vars (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES code_analysis.files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create constants table
CREATE TABLE IF NOT EXISTS code_analysis.constants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES code_analysis.files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create init_functions table
CREATE TABLE IF NOT EXISTS code_analysis.init_functions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES code_analysis.files(id) ON DELETE CASCADE,
    functionality TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create structs table
CREATE TABLE IF NOT EXISTS code_analysis.structs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES code_analysis.files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create struct_fields table
CREATE TABLE IF NOT EXISTS code_analysis.struct_fields (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    struct_id UUID NOT NULL REFERENCES code_analysis.structs(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create methods table
CREATE TABLE IF NOT EXISTS code_analysis.methods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES code_analysis.files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    receiver VARCHAR(255),
    functionality TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create method_params table
CREATE TABLE IF NOT EXISTS code_analysis.method_params (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    method_id UUID NOT NULL REFERENCES code_analysis.methods(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    is_input BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create workflow_steps table
CREATE TABLE IF NOT EXISTS code_analysis.workflow_steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES code_analysis.files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    step_type VARCHAR(100) NOT NULL,
    type_details TEXT,
    description TEXT,
    workflow_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create workflow_step_dependencies table
CREATE TABLE IF NOT EXISTS code_analysis.workflow_step_dependencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_step_id UUID NOT NULL REFERENCES code_analysis.workflow_steps(id) ON DELETE CASCADE,
    dependency VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create workflow_step_variables table
CREATE TABLE IF NOT EXISTS code_analysis.workflow_step_variables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_step_id UUID NOT NULL REFERENCES code_analysis.workflow_steps(id) ON DELETE CASCADE,
    variable_name VARCHAR(255) NOT NULL,
    is_input BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_repositories_owner_name ON code_analysis.repositories(owner, name);
CREATE INDEX IF NOT EXISTS idx_files_repository_id ON code_analysis.files(repository_id);
CREATE INDEX IF NOT EXISTS idx_dependencies_file_id ON code_analysis.dependencies(file_id);
CREATE INDEX IF NOT EXISTS idx_global_vars_file_id ON code_analysis.global_vars(file_id);
CREATE INDEX IF NOT EXISTS idx_constants_file_id ON code_analysis.constants(file_id);
CREATE INDEX IF NOT EXISTS idx_init_functions_file_id ON code_analysis.init_functions(file_id);
CREATE INDEX IF NOT EXISTS idx_structs_file_id ON code_analysis.structs(file_id);
CREATE INDEX IF NOT EXISTS idx_struct_fields_struct_id ON code_analysis.struct_fields(struct_id);
CREATE INDEX IF NOT EXISTS idx_methods_file_id ON code_analysis.methods(file_id);
CREATE INDEX IF NOT EXISTS idx_method_params_method_id ON code_analysis.method_params(method_id);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_file_id ON code_analysis.workflow_steps(file_id);
CREATE INDEX IF NOT EXISTS idx_workflow_step_dependencies_workflow_step_id ON code_analysis.workflow_step_dependencies(workflow_step_id);
CREATE INDEX IF NOT EXISTS idx_workflow_step_variables_workflow_step_id ON code_analysis.workflow_step_variables(workflow_step_id);
