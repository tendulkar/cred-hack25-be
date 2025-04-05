# Code Analysis API

This API allows you to analyze GitHub repositories and extract detailed information about the code structure and functionality.

## Analyze Repository

Analyzes a GitHub repository and extracts detailed information about the code structure and functionality.

### Endpoint

```
POST /api/v1/code/analyze
```

### Request

```json
{
  "repo_url": "https://github.com/username/repo",
  "auth_token": "github_personal_access_token"  // Optional, for private repositories
}
```

### Response

```json
{
  "repo_url": "https://github.com/username/repo",
  "owner": "username",
  "name": "repo",
  "files": [
    {
      "path": "src/main.go",
      "dependencies": [
        "fmt",
        "os",
        "github.com/example/package"
      ],
      "global_vars": [
        {
          "name": "defaultTimeout",
          "type": "int",
          "value": "30"
        }
      ],
      "constants": [
        {
          "name": "MaxRetries",
          "type": "int",
          "value": "3"
        }
      ],
      "init_function": {
        "name": "init",
        "input_params": [],
        "output_params": [],
        "functionality": "Initializes the application by setting up logging and loading configuration."
      },
      "structs": [
        {
          "name": "Config",
          "fields": [
            {
              "name": "Host",
              "type": "string"
            },
            {
              "name": "Port",
              "type": "int"
            }
          ]
        }
      ],
      "methods": [
        {
          "name": "NewConfig",
          "receiver": "",
          "input_params": [
            {
              "name": "path",
              "type": "string"
            }
          ],
          "output_params": [
            {
              "name": "",
              "type": "*Config"
            },
            {
              "name": "",
              "type": "error"
            }
          ],
          "functionality": "Creates a new Config instance by loading from the specified path."
        }
      ],
      "workflow_steps": [
        {
          "name": "Load Configuration",
          "type": "file_operation",
          "type_details": "Read JSON file",
          "description": "Reads configuration from a JSON file",
          "dependencies": ["os", "encoding/json"],
          "input_vars": ["path"],
          "output_vars": ["config", "err"],
          "workflow_name": "Configuration Loading"
        }
      ]
    }
  ]
}
```

## Features

The code analysis API extracts the following information from each file:

1. **Dependencies**: Imports, includes, and other external dependencies.
2. **Global Variables**: Variables defined at the package or module level.
3. **Constants**: Constant values defined in the code.
4. **Init Function**: Information about initialization functions, if present.
5. **Structs**: Information about data structures, including fields and types.
6. **Methods**: Information about methods and functions, including:
   - Input parameters
   - Output parameters
   - Functionality description
7. **Workflow Steps**: Logical steps in the code execution, including:
   - Step name
   - Step type (external system, database, logic, function call)
   - Type details (external system name, database schema, operation name, function call details)
   - Step description
   - Dependencies
   - Input variables/objects
   - Output variables/objects
   - Workflow name

## Post-Order Traversal

The API performs a post-order traversal of the repository files, which means:

1. It processes all files in a directory before processing the directory itself.
2. For each directory, it processes subdirectories before processing files in the current directory.

This ensures that dependencies are analyzed before the files that depend on them.

## Authentication

For private repositories, you can provide a GitHub personal access token in the `auth_token` field of the request. The token should have the `repo` scope to access private repositories.

## Error Handling

The API returns appropriate HTTP status codes and error messages:

- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Invalid or missing authentication token
- `404 Not Found`: Repository not found
- `500 Internal Server Error`: Server-side error

## Example Usage

```bash
curl -X POST http://localhost:8080/api/v1/code/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "repo_url": "https://github.com/username/repo",
    "auth_token": "github_personal_access_token"
  }'
