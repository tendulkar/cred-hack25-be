# Code Analyzer API

The Code Analyzer API provides functionality to analyze Go codebases for symbols, structure, call hierarchies, and references.

## Base URL

```
/api/code-analyzer
```

## Endpoints

### Index Repository

Indexes a GitHub repository, analyzing all Go files and storing the analysis in the database.

**URL**: `/repositories`
**Method**: `POST`
**Auth required**: Yes

#### Request Body

```json
{
  "url": "https://github.com/username/repository"
}
```

#### Success Response

**Code**: `200 OK`
**Content Example**:

```json
{
  "id": 1,
  "url": "https://github.com/username/repository",
  "index_status": "in_progress",
  "message": "Repository indexing started"
}
```

#### Error Responses

**Condition**: Invalid request format or URL is missing.
**Code**: `400 Bad Request`
**Content**:

```json
{
  "error": "URL is required"
}
```

**Condition**: Server error during indexing.
**Code**: `500 Internal Server Error`
**Content**:

```json
{
  "error": "error description"
}
```

### Get Repository Index

Retrieves analysis information for a repository, either for the entire repository or for a specific file. Includes call graph information.

**URL**: `/repositories`
**Method**: `GET`
**Auth required**: Yes

#### Query Parameters

- `url` (required): GitHub repository URL.
- `file_path` (optional): Relative path to a specific file within the repository.
- `include_call_graph` (optional): Set to "true" to include detailed call graph information. Default is "false".

#### Success Response

**Code**: `200 OK`
**Content Example** (for a repository):

```json
{
  "repository": {
    "id": 1,
    "kind": "github",
    "url": "https://github.com/username/repository",
    "name": "repository",
    "owner": "username",
    "local_path": "/path/to/local/clone",
    "last_indexed": "2025-05-01T12:00:00Z",
    "index_status": "completed",
    "index_error": null,
    "created_at": "2025-05-01T11:30:00Z",
    "updated_at": "2025-05-01T12:00:00Z"
  },
  "files": [
    {
      "id": 1,
      "repository_id": 1,
      "file_path": "main.go",
      "package": "main",
      "last_analyzed": "2025-05-01T12:00:00Z",
      "created_at": "2025-05-01T11:30:00Z",
      "updated_at": "2025-05-01T12:00:00Z"
    }
  ],
  "call_graph": {
    "nodes": [
      {
        "id": "main.main",
        "package": "main",
        "function": "main",
        "file_path": "main.go"
      },
      {
        "id": "fmt.Println",
        "package": "fmt",
        "function": "Println",
        "is_external": true
      }
    ],
    "edges": [
      {
        "source": "main.main",
        "target": "fmt.Println",
        "line": 12,
        "count": 1
      }
    ]
  }
}
```

**Content Example** (for a specific file):

```json
{
  "repository": { /* Repository details */ },
  "files": [
    { /* File details */ }
  ],
  "functions": [
    {
      "id": 1,
      "repository_id": 1,
      "file_id": 1,
      "name": "main",
      "kind": "function",
      "exported": true,
      "parameters": "[]",
      "results": "[]",
      "code_block": "func main() {\n  // Code here\n}",
      "line": 10,
      "calls": "[{\"callee\":\"fmt.Println\"}]",
      "called_by": "[]",
      "references": "[]",
      "created_at": "2025-05-01T11:30:00Z",
      "updated_at": "2025-05-01T12:00:00Z"
    }
  ],
  "symbols": [
    {
      "id": 1,
      "repository_id": 1,
      "file_id": 1,
      "name": "VERSION",
      "kind": "constant",
      "type": "string",
      "value": "\"1.0.0\"",
      "exported": true,
      "fields": null,
      "methods": null,
      "line": 5,
      "references": "[{\"ref_type\":\"usage\",\"line\":15}]",
      "created_at": "2025-05-01T11:30:00Z",
      "updated_at": "2025-05-01T12:00:00Z"
    }
  ]
}
```

#### Error Responses

**Condition**: URL parameter is missing.
**Code**: `400 Bad Request`
**Content**:

```json
{
  "error": "URL is required"
}
```

**Condition**: Repository not found or server error.
**Code**: `500 Internal Server Error`
**Content**:

```json
{
  "error": "error description"
}
```

### Analyze Single File

Analyzes a single Go file without storing the results in the database.

**URL**: `/analyze-file`
**Method**: `POST`
**Auth required**: Yes

#### Request Body

```json
{
  "file_path": "/path/to/file.go"
}
```

#### Success Response

**Code**: `200 OK`
**Content**: Detailed file analysis in JSON format (similar to the repository file analysis).

#### Error Responses

**Condition**: Invalid request format or file path is missing.
**Code**: `400 Bad Request`
**Content**:

```json
{
  "error": "Invalid request format"
}
```

**Condition**: File not found or server error.
**Code**: `500 Internal Server Error`
**Content**:

```json
{
  "error": "error description"
}
```

## Models

### Core Models

#### Repository

```json
{
  "id": 1,
  "kind": "github",
  "url": "https://github.com/username/repository",
  "name": "repository",
  "owner": "username",
  "local_path": "/path/to/local/clone",
  "last_indexed": "2025-05-01T12:00:00Z",
  "index_status": "completed",
  "index_error": null,
  "created_at": "2025-05-01T11:30:00Z",
  "updated_at": "2025-05-01T12:00:00Z"
}
```

#### RepositoryFile

```json
{
  "id": 1,
  "repository_id": 1,
  "file_path": "main.go",
  "package": "main",
  "last_analyzed": "2025-05-01T12:00:00Z",
  "created_at": "2025-05-01T11:30:00Z",
  "updated_at": "2025-05-01T12:00:00Z"
}
```

#### RepositoryFunction

```json
{
  "id": 1,
  "repository_id": 1,
  "file_id": 1,
  "name": "main",
  "kind": "function",
  "receiver": "",
  "exported": true,
  "parameters": "[]",
  "results": "[]",
  "code_block": "func main() {\n  // Code here\n}",
  "line": 10,
  "created_at": "2025-05-01T11:30:00Z",
  "updated_at": "2025-05-01T12:00:00Z"
}
```

#### RepositorySymbol

```json
{
  "id": 1,
  "repository_id": 1,
  "file_id": 1,
  "name": "VERSION",
  "kind": "constant",
  "type": "string",
  "value": "\"1.0.0\"",
  "exported": true,
  "fields": null,
  "methods": null,
  "line": 5,
  "created_at": "2025-05-01T11:30:00Z",
  "updated_at": "2025-05-01T12:00:00Z"
}
```

### Relationship Models

#### FunctionCall

```json
{
  "id": 1,
  "caller_id": 1,
  "callee_name": "fmt.Println",
  "callee_package": "fmt",
  "callee_id": null,
  "line": 12,
  "parameters": "[\"Hello, World!\"]",
  "created_at": "2025-05-01T11:30:00Z",
  "updated_at": "2025-05-01T12:00:00Z"
}
```

#### FunctionReference

```json
{
  "id": 1,
  "function_id": 1,
  "reference_type": "declaration",
  "file_id": 1,
  "line": 10,
  "column_position": 1,
  "context": "func main() {",
  "created_at": "2025-05-01T11:30:00Z",
  "updated_at": "2025-05-01T12:00:00Z"
}
```

#### FunctionStatement

```json
{
  "id": 1,
  "function_id": 1,
  "statement_type": "if",
  "text": "if err != nil {",
  "line": 15,
  "conditions": "{\"left\":\"err\",\"operator\":\"!=\",\"right\":\"nil\"}",
  "variables": "[\"err\"]",
  "calls": "[]",
  "parent_statement_id": null,
  "created_at": "2025-05-01T11:30:00Z",
  "updated_at": "2025-05-01T12:00:00Z"
}
```

#### SymbolReference

```json
{
  "id": 1,
  "symbol_id": 1,
  "reference_type": "usage",
  "file_id": 1,
  "line": 15,
  "column_position": 10,
  "context": "fmt.Println(VERSION)",
  "created_at": "2025-05-01T11:30:00Z",
  "updated_at": "2025-05-01T12:00:00Z"
}
```

### Call Graph Models

#### CallGraphNode

```json
{
  "id": "main.main",
  "package": "main",
  "function": "main", 
  "receiver": "",
  "file_path": "main.go",
  "line": 10,
  "is_external": false
}
```

The `id` field is a unique identifier for the function, typically in the format `package.function` or `package.receiver.function` for methods.

#### CallGraphEdge

```json
{
  "source": "main.main",
  "target": "fmt.Println",
  "line": 12,
  "parameters": "[\"Hello, World!\"]",
  "count": 1
}
```

The edge represents a function call, where `source` is the caller function ID and `target` is the callee function ID.

#### CompleteCallGraph

```json
{
  "nodes": [
    {
      "id": "main.main",
      "package": "main",
      "function": "main",
      "file_path": "main.go",
      "line": 10,
      "is_external": false
    },
    {
      "id": "fmt.Println",
      "package": "fmt",
      "function": "Println",
      "is_external": true
    }
  ],
  "edges": [
    {
      "source": "main.main",
      "target": "fmt.Println",
      "line": 12,
      "parameters": "[\"Hello, World!\"]",
      "count": 1
    }
  ]
}
```

### Extended Response Models

For detailed queries, the API may return extended models that include related data:

#### ExtendedRepositoryFunction

```json
{
  "function": {
    "id": 1,
    "repository_id": 1,
    "file_id": 1,
    "name": "main",
    "kind": "function",
    "receiver": "",
    "exported": true,
    "parameters": "[]",
    "results": "[]",
    "code_block": "func main() {\n  // Code here\n}",
    "line": 10,
    "created_at": "2025-05-01T11:30:00Z", 
    "updated_at": "2025-05-01T12:00:00Z"
  },
  "calls": [
    {
      "id": 1,
      "caller_id": 1,
      "callee_name": "fmt.Println",
      "callee_package": "fmt",
      "callee_id": null,
      "line": 12,
      "parameters": "[\"Hello, World!\"]",
      "created_at": "2025-05-01T11:30:00Z",
      "updated_at": "2025-05-01T12:00:00Z"
    }
  ],
  "references": [
    {
      "id": 1,
      "function_id": 1,
      "reference_type": "declaration",
      "file_id": 1,
      "line": 10,
      "column_position": 1,
      "context": "func main() {",
      "created_at": "2025-05-01T11:30:00Z",
      "updated_at": "2025-05-01T12:00:00Z"
    }
  ],
  "statements": [
    {
      "id": 1,
      "function_id": 1,
      "statement_type": "if",
      "text": "if err != nil {",
      "line": 15,
      "conditions": "{\"left\":\"err\",\"operator\":\"!=\",\"right\":\"nil\"}",
      "variables": "[\"err\"]",
      "calls": "[]",
      "parent_statement_id": null,
      "created_at": "2025-05-01T11:30:00Z",
      "updated_at": "2025-05-01T12:00:00Z"
    }
  ]
}
```

#### ExtendedRepositorySymbol

```json
{
  "symbol": {
    "id": 1,
    "repository_id": 1,
    "file_id": 1,
    "name": "VERSION",
    "kind": "constant",
    "type": "string",
    "value": "\"1.0.0\"",
    "exported": true,
    "fields": null,
    "methods": null,
    "line": 5,
    "created_at": "2025-05-01T11:30:00Z",
    "updated_at": "2025-05-01T12:00:00Z"
  },
  "references": [
    {
      "id": 1,
      "symbol_id": 1,
      "reference_type": "usage",
      "file_id": 1,
      "line": 15,
      "column_position": 10,
      "context": "fmt.Println(VERSION)",
      "created_at": "2025-05-01T11:30:00Z",
      "updated_at": "2025-05-01T12:00:00Z"
    }
  ]
}
```
