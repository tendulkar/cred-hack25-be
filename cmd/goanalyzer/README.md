# Go Code Analyzer

A command-line tool for analyzing Go source code files and extracting information about symbols, call hierarchies, and references.

## Features

- Analyze individual Go files or entire directories
- Extract symbols (packages, imports, constants, variables, types, functions, methods, etc.)
- Identify call hierarchies between functions
- Detect references to symbols across files
- Output in JSON or text format

## Usage

```bash
# Analyze a single file
go run cmd/goanalyzer/main.go -path=/path/to/file.go

# Analyze a directory (non-recursive)
go run cmd/goanalyzer/main.go -path=/path/to/directory

# Analyze a directory recursively
go run cmd/goanalyzer/main.go -path=/path/to/directory -recursive=true

# Output in text format
go run cmd/goanalyzer/main.go -path=/path/to/file.go -format=text

# Save output to a file
go run cmd/goanalyzer/main.go -path=/path/to/file.go -output=analysis.json
```

## Command-line Options

- `-path`: Path to a Go file or directory (required)
- `-recursive`: Recursively analyze directories (default: false)
- `-format`: Output format - "json" or "text" (default: "json")
- `-output`: Output file path (default: stdout)

## Output Format

### JSON Format

The JSON output includes:
- File path
- List of symbols with their properties:
  - Name
  - Kind (package, import, const, var, type, func, struct, interface, etc.)
  - Line number
  - Exported status
  - Type information
  - Fields (for structs)
  - Methods (for types)
  - Parameters and results (for functions)
  - Function calls (for functions)

### Text Format

The text output is organized by symbol kind and includes:
- File path
- Symbols grouped by kind (PACKAGE, IMPORT, CONST, VAR, TYPE, FUNC, etc.)
- Call hierarchy section showing which functions call other functions
- Exported symbols are marked with an asterisk (*)

## Example

Analyzing a simple Go file:

```go
package main

import "fmt"

const Version = "1.0.0"

func main() {
    fmt.Println("Hello, world!")
    greet("User")
}

func greet(name string) {
    fmt.Printf("Hello, %s!\n", name)
}
```

Will produce output showing:
- Package: main
- Import: fmt
- Constant: Version
- Functions: main, greet
- Call hierarchy: main calls fmt.Println and greet, greet calls fmt.Printf

## Integration

This tool can be integrated into CI/CD pipelines to:
- Generate documentation
- Analyze code complexity
- Track dependencies between components
- Validate architectural constraints
