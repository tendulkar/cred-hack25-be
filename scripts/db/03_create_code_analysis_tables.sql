-- Connect to the database
\c code_analyser

-- Create repositories table
CREATE TABLE IF NOT EXISTS repositories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_url VARCHAR(255) UNIQUE NOT NULL,
    owner VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    analyzed_at TIMESTAMP WITH TIME ZONE
);

-- Create files table
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    path VARCHAR(1000) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(repository_id, path)
);

-- Create dependencies table
CREATE TABLE IF NOT EXISTS dependencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create global_vars table
CREATE TABLE IF NOT EXISTS global_vars (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create constants table
CREATE TABLE IF NOT EXISTS constants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create init_functions table
CREATE TABLE IF NOT EXISTS init_functions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    functionality TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create structs table
CREATE TABLE IF NOT EXISTS structs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create struct_fields table
CREATE TABLE IF NOT EXISTS struct_fields (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    struct_id UUID NOT NULL REFERENCES structs(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create methods table
CREATE TABLE IF NOT EXISTS methods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    receiver VARCHAR(255),
    functionality TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create method_params table
CREATE TABLE IF NOT EXISTS method_params (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    method_id UUID NOT NULL REFERENCES methods(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    is_input BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create workflow_steps table
CREATE TABLE IF NOT EXISTS workflow_steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    step_type VARCHAR(100) NOT NULL,
    type_details TEXT,
    description TEXT,
    workflow_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create workflow_step_dependencies table
CREATE TABLE IF NOT EXISTS workflow_step_dependencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_step_id UUID NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    dependency VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create workflow_step_variables table
CREATE TABLE IF NOT EXISTS workflow_step_variables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_step_id UUID NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    variable_name VARCHAR(255) NOT NULL,
    is_input BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_repositories_owner_name ON repositories(owner, name);
CREATE INDEX IF NOT EXISTS idx_files_repository_id ON files(repository_id);
CREATE INDEX IF NOT EXISTS idx_dependencies_file_id ON dependencies(file_id);
CREATE INDEX IF NOT EXISTS idx_global_vars_file_id ON global_vars(file_id);
CREATE INDEX IF NOT EXISTS idx_constants_file_id ON constants(file_id);
CREATE INDEX IF NOT EXISTS idx_init_functions_file_id ON init_functions(file_id);
CREATE INDEX IF NOT EXISTS idx_structs_file_id ON structs(file_id);
CREATE INDEX IF NOT EXISTS idx_struct_fields_struct_id ON struct_fields(struct_id);
CREATE INDEX IF NOT EXISTS idx_methods_file_id ON methods(file_id);
CREATE INDEX IF NOT EXISTS idx_method_params_method_id ON method_params(method_id);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_file_id ON workflow_steps(file_id);
CREATE INDEX IF NOT EXISTS idx_workflow_step_dependencies_workflow_step_id ON workflow_step_dependencies(workflow_step_id);
CREATE INDEX IF NOT EXISTS idx_workflow_step_variables_workflow_step_id ON workflow_step_variables(workflow_step_id);

-- Grant privileges to the user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO code_analyser_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO code_analyser_user;
