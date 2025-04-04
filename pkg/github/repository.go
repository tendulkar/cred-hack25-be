package github

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// GetFile returns a file from the repository by path
func (r *Repository) GetFile(filePath string) (*File, error) {
	file, exists := r.Files[filePath]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}
	return file, nil
}

// ListDirectory returns a list of files and directories in the given directory
func (r *Repository) ListDirectory(dirPath string) ([]string, error) {
	// Normalize empty path to root
	if dirPath == "" || dirPath == "." {
		dirPath = ""
	}

	// Check if directory exists
	if dirPath != "" {
		if _, exists := r.Files[dirPath]; !exists {
			return nil, fmt.Errorf("directory not found: %s", dirPath)
		}
	}

	// Get files in directory
	files, exists := r.FileTree[dirPath]
	if !exists {
		return nil, fmt.Errorf("directory not found: %s", dirPath)
	}

	return files, nil
}

// FindFiles returns a list of files that match the given pattern
func (r *Repository) FindFiles(pattern string) ([]string, error) {
	var matches []string

	for filePath := range r.Files {
		matched, err := filepath.Match(pattern, path.Base(filePath))
		if err != nil {
			return nil, err
		}
		if matched {
			matches = append(matches, filePath)
		}
	}

	return matches, nil
}

// SearchContent returns a list of files that contain the given text
func (r *Repository) SearchContent(text string) []string {
	var matches []string

	for filePath, file := range r.Files {
		if file.Type == "file" && strings.Contains(file.Content, text) {
			matches = append(matches, filePath)
		}
	}

	return matches
}

// GetDirectoryTree returns a tree representation of the repository
func (r *Repository) GetDirectoryTree() map[string][]string {
	return r.FileTree
}

// GetFileContent returns the content of a file
func (r *Repository) GetFileContent(filePath string) (string, error) {
	file, err := r.GetFile(filePath)
	if err != nil {
		return "", err
	}

	if file.Type != "file" {
		return "", fmt.Errorf("not a file: %s", filePath)
	}

	return file.Content, nil
}

// IsDirectory checks if the given path is a directory
func (r *Repository) IsDirectory(dirPath string) bool {
	file, exists := r.Files[dirPath]
	return exists && file.Type == "dir"
}

// IsFile checks if the given path is a file
func (r *Repository) IsFile(filePath string) bool {
	file, exists := r.Files[filePath]
	return exists && file.Type == "file"
}

// GetAllFiles returns a list of all files in the repository
func (r *Repository) GetAllFiles() []*File {
	var files []*File
	for _, file := range r.Files {
		files = append(files, file)
	}
	return files
}

// GetAllFilePaths returns a list of all file paths in the repository
func (r *Repository) GetAllFilePaths() []string {
	var paths []string
	for path, file := range r.Files {
		if file.Type == "file" {
			paths = append(paths, path)
		}
	}
	return paths
}
