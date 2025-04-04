package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"cred.com/hack25/backend/pkg/github"
	"cred.com/hack25/backend/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init(logger.InfoLevel, "")

	// Parse command-line flags
	repoURL := flag.String("repo", "", "GitHub repository URL (required)")
	listDir := flag.String("ls", "", "List files in directory")
	getFile := flag.String("cat", "", "Get file content")
	findPattern := flag.String("find", "", "Find files matching pattern")
	searchText := flag.String("search", "", "Search for files containing text")
	flag.Parse()

	// Check if repository URL is provided
	if *repoURL == "" {
		fmt.Println("Error: GitHub repository URL is required")
		fmt.Println("Usage: github-client -repo=https://github.com/user/repo [options]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Create GitHub client
	client := github.NewClient()

	// Fetch repository
	logger.Infof("Fetching repository: %s", *repoURL)
	repo, err := client.FetchRepository(*repoURL)
	if err != nil {
		logger.Errorf("Failed to fetch repository: %v", err)
		os.Exit(1)
	}

	logger.Infof("Successfully fetched repository: %s/%s", repo.Owner, repo.Name)

	// Process commands
	if *listDir != "" {
		listDirectory(repo, *listDir)
	} else if *getFile != "" {
		getFileContent(repo, *getFile)
	} else if *findPattern != "" {
		findFiles(repo, *findPattern)
	} else if *searchText != "" {
		searchFiles(repo, *searchText)
	} else {
		// Default: show repository structure
		showRepositoryStructure(repo)
	}
}

func listDirectory(repo *github.Repository, dirPath string) {
	files, err := repo.ListDirectory(dirPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if dirPath == "" {
		dirPath = "/"
	}

	fmt.Printf("Contents of %s:\n", dirPath)
	for _, file := range files {
		if repo.IsDirectory(file) {
			fmt.Printf("📁 %s/\n", file)
		} else {
			fmt.Printf("📄 %s\n", file)
		}
	}
}

func getFileContent(repo *github.Repository, filePath string) {
	content, err := repo.GetFileContent(filePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Content of %s:\n", filePath)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(content)
	fmt.Println(strings.Repeat("-", 80))
}

func findFiles(repo *github.Repository, pattern string) {
	files, err := repo.FindFiles(pattern)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Files matching pattern '%s':\n", pattern)
	for _, file := range files {
		fmt.Println(file)
	}
	fmt.Printf("Found %d files\n", len(files))
}

func searchFiles(repo *github.Repository, text string) {
	files := repo.SearchContent(text)

	fmt.Printf("Files containing '%s':\n", text)
	for _, file := range files {
		fmt.Println(file)
	}
	fmt.Printf("Found %d files\n", len(files))
}

func showRepositoryStructure(repo *github.Repository) {
	fmt.Printf("Repository: %s/%s\n\n", repo.Owner, repo.Name)

	// Get directory tree
	tree := repo.GetDirectoryTree()

	// Print root directory
	fmt.Println("Repository structure:")
	printDirectoryTree(repo, tree, "", "", 0, 2) // Max depth of 2
}

func printDirectoryTree(repo *github.Repository, tree map[string][]string, dirPath string, prefix string, depth, maxDepth int) {
	if depth > maxDepth {
		fmt.Printf("%s...\n", prefix)
		return
	}

	files, exists := tree[dirPath]
	if !exists {
		return
	}

	for i, file := range files {
		isLast := i == len(files)-1
		var newPrefix string

		if isLast {
			fmt.Printf("%s└── ", prefix)
			newPrefix = prefix + "    "
		} else {
			fmt.Printf("%s├── ", prefix)
			newPrefix = prefix + "│   "
		}

		if repo.IsDirectory(file) {
			fmt.Printf("📁 %s/\n", file)
			printDirectoryTree(repo, tree, file, newPrefix, depth+1, maxDepth)
		} else {
			fmt.Printf("📄 %s\n", file)
		}
	}
}
