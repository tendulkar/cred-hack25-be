# GitHub Client

A Go package for fetching and working with GitHub repositories.

## Features

- Fetch code from a GitHub repository URL
- Build an in-memory logical git repository by file paths
- Navigate the repository structure
- Search for files by pattern or content
- Get file contents

## Usage

### Basic Usage

```go
package main

import (
	"fmt"
	"log"

	"cred.com/hack25/backend/pkg/github"
)

func main() {
	// Create a new GitHub client
	client := github.NewClient()

	// Fetch a repository
	repo, err := client.FetchRepository("https://github.com/username/repo")
	if err != nil {
		log.Fatalf("Failed to fetch repository: %v", err)
	}

	// Print repository information
	fmt.Printf("Repository: %s/%s\n", repo.Owner, repo.Name)

	// List files in the root directory
	files, err := repo.ListDirectory("")
	if err != nil {
		log.Fatalf("Failed to list directory: %v", err)
	}

	fmt.Println("Files in root directory:")
	for _, file := range files {
		fmt.Println("-", file)
	}
}
```

### Working with Files

```go
// Get content of a specific file
content, err := repo.GetFileContent("README.md")
if err != nil {
	log.Fatalf("Failed to get file content: %v", err)
}
fmt.Println("README.md content:", content)

// Check if a path is a directory
isDir := repo.IsDirectory("src")
fmt.Println("Is 'src' a directory?", isDir)

// Check if a path is a file
isFile := repo.IsFile("README.md")
fmt.Println("Is 'README.md' a file?", isFile)
```

### Searching and Finding Files

```go
// Find all Go files
goFiles, err := repo.FindFiles("*.go")
if err != nil {
	log.Fatalf("Failed to find files: %v", err)
}
fmt.Printf("Found %d Go files\n", len(goFiles))

// Search for files containing specific text
matches := repo.SearchContent("fmt.Println")
fmt.Printf("Found %d files containing 'fmt.Println'\n", len(matches))
```

### Getting Repository Structure

```go
// Get directory tree
tree := repo.GetDirectoryTree()
fmt.Println("Repository structure:", tree)

// Get all file paths
paths := repo.GetAllFilePaths()
fmt.Printf("Repository has %d files\n", len(paths))
```

## Command-Line Tool

A command-line tool is also provided to demonstrate the GitHub client functionality. You can find it in `cmd/github-client/main.go`.

### Building the Command-Line Tool

```bash
go build -o github-client cmd/github-client/main.go
```

### Using the Command-Line Tool

```bash
# Show repository structure
./github-client -repo=https://github.com/username/repo

# List files in a directory
./github-client -repo=https://github.com/username/repo -ls=src

# Get file content
./github-client -repo=https://github.com/username/repo -cat=README.md

# Find files matching a pattern
./github-client -repo=https://github.com/username/repo -find="*.go"

# Search for files containing text
./github-client -repo=https://github.com/username/repo -search="fmt.Println"
```

## API Reference

### Client Interface

```go
type Client interface {
	// FetchRepository fetches the repository content from the given URL
	FetchRepository(repoURL string) (*Repository, error)
}
```

### Repository Methods

```go
// GetFile returns a file from the repository by path
func (r *Repository) GetFile(filePath string) (*File, error)

// ListDirectory returns a list of files and directories in the given directory
func (r *Repository) ListDirectory(dirPath string) ([]string, error)

// FindFiles returns a list of files that match the given pattern
func (r *Repository) FindFiles(pattern string) ([]string, error)

// SearchContent returns a list of files that contain the given text
func (r *Repository) SearchContent(text string) []string

// GetDirectoryTree returns a tree representation of the repository
func (r *Repository) GetDirectoryTree() map[string][]string

// GetFileContent returns the content of a file
func (r *Repository) GetFileContent(filePath string) (string, error)

// IsDirectory checks if the given path is a directory
func (r *Repository) IsDirectory(dirPath string) bool

// IsFile checks if the given path is a file
func (r *Repository) IsFile(filePath string) bool

// GetAllFiles returns a list of all files in the repository
func (r *Repository) GetAllFiles() []*File

// GetAllFilePaths returns a list of all file paths in the repository
func (r *Repository) GetAllFilePaths() []string
```

## Limitations

- The GitHub API has rate limits, so fetching large repositories may be limited
- Binary files are not fully supported
- Authentication is not currently implemented, so only public repositories can be accessed
