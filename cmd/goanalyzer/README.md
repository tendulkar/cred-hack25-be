# Go Code Analyzer

A comprehensive tool for analyzing Go source code to identify symbols, track call hierarchies, and find references.

## Features

- **Complete Symbol Identification**: Analyzes Go files to identify and categorize all symbols
  - Packages
  - Imports
  - Constants
  - Variables
  - Types (including structs and interfaces)
  - Functions and methods
  - Fields and parameters

- **Deep Code Analysis**:
  - Statement-level analysis with AST parsing
  - Call hierarchy tracking (which functions call which)
  - Reference tracking (where symbols are used)
  - Exported/unexported status of symbols

- **Flexible Output Formats**:
  - Human-readable text output with customizable detail levels
  - JSON output for machine processing and integration with other tools

## Installation

```bash
# From the project root
cd cmd/goanalyzer
go build
```

## Usage

```bash
# Analyze a single file
./goanalyzer -file=/path/to/file.go

# Analyze an entire directory
./goanalyzer -dir=/path/to/directory

# Show function code blocks
./goanalyzer -file=/path/to/file.go -code

# Show call hierarchy
./goanalyzer -file=/path/to/file.go -calls

# Show symbol references
./goanalyzer -file=/path/to/file.go -refs

# Show statement analysis
./goanalyzer -file=/path/to/file.go -statements

# Focus on a specific function
./goanalyzer -file=/path/to/file.go -function=FunctionName -statements

# Output as JSON
./goanalyzer -file=/path/to/file.go -json
```

## Examples

### Symbol Listing

Running the analyzer on a Go file will list all symbols found:

```
=== File: /path/to/file.go ===
Package: main

Imports:
  fmt: fmt
  os: os

Constants:
  MaxRetries: int = 3 (exported: true)

Variables:
  defaultTimeout: time.Duration = 30 * time.Second (exported: false)

Functions:
  Function: main (exported: true)
  Method: Process on *Processor (exported: true)
```

### Call Hierarchy

Using the `-calls` flag shows which functions call other functions:

```
Functions:
  Function: main (exported: true)
    Calls:
      -> fmt.Println (line 15)
      -> NewProcessor (line 17)
      -> processor.Process (line 18)
```

### Symbol References

Using the `-refs` flag shows where symbols are used:

```
References:
  main.Processor:
    1 declaration(s), 3 usage(s)
    - Used at line 17
    - Used at line 25
    - Used at line 42
```

## Architecture

The analyzer is built using Go's standard library packages:
- `go/ast`: For Abstract Syntax Tree parsing
- `go/parser`: For parsing Go source code
- `go/token`: For token operations

The code is structured into three main packages:
1. **models**: Data structures for representing code elements
2. **analyzer**: Core analysis engine with AST parsing
3. **cmd/goanalyzer**: Command-line interface
