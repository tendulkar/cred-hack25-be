package github_test

import (
	"fmt"
	"log"

	"cred.com/hack25/backend/pkg/github"
)

func Example() {
	// Create a new GitHub client
	client := github.NewClient("")

	// Fetch a repository
	repo, err := client.FetchRepository("https://github.com/golang/go", "")
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

	// Find all Go files
	goFiles, err := repo.FindFiles("*.go")
	if err != nil {
		log.Fatalf("Failed to find files: %v", err)
	}

	fmt.Printf("Found %d Go files\n", len(goFiles))
	if len(goFiles) > 5 {
		fmt.Println("First 5 Go files:")
		for i := 0; i < 5; i++ {
			fmt.Println("-", goFiles[i])
		}
	}

	// Search for files containing "fmt.Println"
	matches := repo.SearchContent("fmt.Println")
	fmt.Printf("Found %d files containing 'fmt.Println'\n", len(matches))
	if len(matches) > 5 {
		fmt.Println("First 5 matches:")
		for i := 0; i < 5; i++ {
			fmt.Println("-", matches[i])
		}
	}

	// Output:
	// (Output will vary based on the repository content)
}

func ExampleRepository_GetFileContent() {
	// Create a new GitHub client
	client := github.NewClient("")

	// Fetch a repository
	repo, err := client.FetchRepository("https://github.com/golang/example", "")
	if err != nil {
		log.Fatalf("Failed to fetch repository: %v", err)
	}

	// Get content of a specific file
	content, err := repo.GetFileContent("README.md")
	if err != nil {
		log.Fatalf("Failed to get file content: %v", err)
	}

	fmt.Printf("README.md content (first 100 chars): %s\n", content[:min(100, len(content))])

	// Output:
	// (Output will vary based on the repository content)
}

func ExampleRepository_GetDirectoryTree() {
	// Create a new GitHub client
	client := github.NewClient("")

	// Fetch a small repository
	repo, err := client.FetchRepository("https://github.com/golang/example", "")
	if err != nil {
		log.Fatalf("Failed to fetch repository: %v", err)
	}

	// Get directory tree
	tree := repo.GetDirectoryTree()

	// Print the tree structure (limited to root and first level)
	fmt.Println("Repository structure:")
	fmt.Println("Root:")
	for _, file := range tree[""] {
		fmt.Println("-", file)

		// If it's a directory, print its contents
		if repo.IsDirectory(file) {
			fmt.Printf("  %s/:\n", file)
			subFiles, err := repo.ListDirectory(file)
			if err != nil {
				continue
			}
			for _, subFile := range subFiles {
				fmt.Printf("  - %s\n", subFile)
			}
		}
	}

	// Output:
	// (Output will vary based on the repository content)
}

// Helper function for min of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
