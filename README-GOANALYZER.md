# Go Code Analyzer with Repository Indexing API

A comprehensive tool for analyzing Go source code that identifies symbols, tracks call hierarchies, analyzes references, and parses statements with AST (Abstract Syntax Tree) processing. The system includes both a CLI tool and a REST API with database persistence.

## Features

### Core Analysis Features

- **Complete Symbol Identification**
  - Packages, imports, constants, variables
  - Functions and methods with parameters and return types
  - Structs, interfaces, and type definitions
  - Exported/unexported status tracking

- **Call Hierarchy Analysis**
  - Track which functions call other functions
  - Record caller-callee relationships with line information
  - Store parameters passed in calls

- **Reference Tracking**
  - Find all occurrences of symbols throughout the codebase
  - Differentiate between declarations, usages, and modifications
  - Include context and position information

- **Statement-Level Analysis**
  - Parse Go's Abstract Syntax Tree for deeper code understanding
  - Analyze control structures (if, for, switch statements)
  - Track variable usage and modifications within statements
  - Build hierarchical statement trees (parent-child relationships)

### Repository Management

- **GitHub Repository Integration**
  - Clone and analyze any GitHub repository
  - Track repository analysis status
  - Store the entire codebase structure

- **Persistent Storage**
  - Normalized database schema for efficient querying
  - Separate tables for different entity types
  - Complete referential integrity

- **REST API**
  - Endpoints for analyzing repositories and individual files
  - Query analyzed data with filtering options
  - JSON response format

## Architecture

### Database Schema

The system uses a normalized database schema to store analysis results efficiently:

1. **Core Tables**
   - `repositories`: Stores repository information
   - `repository_files`: Tracks individual files in repositories
   - `repository_functions`: Stores function/method definitions
   - `repository_symbols`: Stores other symbol types (variables, constants, etc.)

2. **Relationship Tables**
   - `function_calls`: Links functions to the functions they call
   - `function_references`: Tracks references to functions
   - `function_statements`: Stores statement analysis within functions
   - `symbol_references`: Tracks references to symbols

### Component Structure

The project is organized into several components:

1. **Command-line Tool** (`cmd/goanalyzer`)
   - Direct Go file/directory analysis
   - Flexible display options

2. **Core Analyzer** (`pkg/goanalyzer`)
   - AST parsing and analysis
   - Symbol extraction
   - Call hierarchy tracking
   - Reference analysis
   - Statement parsing

3. **API Layer**
   - Models: Database entity definitions
   - Repository: Database access logic
   - Service: Business logic
   - Handlers: REST API endpoints

## API Usage

The API provides several endpoints:

### 1. Index a Repository

```
POST /api/code-analyzer/repositories
```

Body:
```json
{
  "url": "https://github.com/username/repository"
}
```

Response:
```json
{
  "id": 1,
  "url": "https://github.com/username/repository",
  "index_status": "in_progress",
  "message": "Repository indexing started"
}
```

### 2. Get Repository Analysis

```
GET /api/code-analyzer/repositories?url=https://github.com/username/repository
```

Response contains repository information, files, and optionally detailed function and symbol information.

### 3. Get File Analysis

```
GET /api/code-analyzer/repositories?url=https://github.com/username/repository&file_path=path/to/file.go
```

Response includes detailed information about a specific file, including all functions, symbols, statements, and references.

### 4. Analyze Single File

```
POST /api/code-analyzer/analyze-file
```

Body:
```json
{
  "file_path": "/path/to/file.go" 
}
```

Response includes complete analysis of the file without storing in the database.

## CLI Usage

The command-line tool provides direct file analysis capabilities:

```bash
# Analyze a file
./goanalyzer -file=/path/to/file.go

# Analyze a directory
./goanalyzer -dir=/path/to/directory

# Show code blocks
./goanalyzer -file=/path/to/file.go -code

# Show call hierarchy
./goanalyzer -file=/path/to/file.go -calls

# Show references
./goanalyzer -file=/path/to/file.go -refs

# Show statement analysis
./goanalyzer -file=/path/to/file.go -statements
```

## Database Setup

```bash
# Run the SQL script to create necessary tables
psql -U username -d database -f scripts/db/05_create_code_analyzer_tables.sql
```

## Future Enhancements

- Support for additional repository providers (GitLab, Bitbucket)
- Integration with CI/CD pipelines for automated code analysis
- Visualization of call graphs and dependency relationships
- Cross-repository reference tracking
- Support for additional languages beyond Go
