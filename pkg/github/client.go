package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
)

// Client is the interface for interacting with GitHub repositories
type Client interface {
	// FetchRepository fetches the repository content from the given URL
	FetchRepository(repoURL string) (*Repository, error)
}

// HTTPClient is the implementation of the Client interface
type HTTPClient struct {
	httpClient *http.Client
}

// NewClient creates a new GitHub client
func NewClient() Client {
	return &HTTPClient{
		httpClient: &http.Client{},
	}
}

// Repository represents a GitHub repository
type Repository struct {
	Owner    string
	Name     string
	Files    map[string]*File
	FileTree map[string][]string // Directory path -> list of files/dirs
}

// File represents a file in a GitHub repository
type File struct {
	Path    string
	Content string
	Type    string // "file" or "dir"
}

// repoURLPattern is a regex pattern to extract owner and repo name from GitHub URLs
var repoURLPattern = regexp.MustCompile(`github\.com[:/]([^/]+)/([^/]+?)(?:\.git)?$`)

// FetchRepository fetches the repository content from the given URL
func (c *HTTPClient) FetchRepository(repoURL string) (*Repository, error) {
	owner, name, err := parseRepoURL(repoURL)
	if err != nil {
		return nil, err
	}

	repo := &Repository{
		Owner:    owner,
		Name:     name,
		Files:    make(map[string]*File),
		FileTree: make(map[string][]string),
	}

	// Initialize root directory
	repo.FileTree[""] = []string{}

	// Fetch repository contents recursively
	err = c.fetchContents(repo, "", owner, name)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// parseRepoURL extracts the owner and repository name from a GitHub URL
func parseRepoURL(repoURL string) (string, string, error) {
	matches := repoURLPattern.FindStringSubmatch(repoURL)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid GitHub repository URL: %s", repoURL)
	}
	return matches[1], matches[2], nil
}

// fetchContents fetches the contents of a directory in the repository
func (c *HTTPClient) fetchContents(repo *Repository, dirPath, owner, name string) error {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, name, dirPath)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}

	var contents []struct {
		Name        string `json:"name"`
		Path        string `json:"path"`
		Type        string `json:"type"`
		DownloadURL string `json:"download_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return err
	}

	// Process each item in the directory
	for _, item := range contents {
		filePath := item.Path

		// Add to file tree
		dir := path.Dir(filePath)
		if dir == "." {
			dir = ""
		}

		if _, exists := repo.FileTree[dir]; !exists {
			repo.FileTree[dir] = []string{}
		}
		repo.FileTree[dir] = append(repo.FileTree[dir], filePath)

		// Process based on type
		if item.Type == "dir" {
			// Create directory entry
			repo.Files[filePath] = &File{
				Path: filePath,
				Type: "dir",
			}

			// Initialize directory in file tree
			repo.FileTree[filePath] = []string{}

			// Recursively fetch contents of subdirectory
			if err := c.fetchContents(repo, filePath, owner, name); err != nil {
				return err
			}
		} else if item.Type == "file" {
			// Fetch file content
			if err := c.fetchFileContent(repo, filePath, item.DownloadURL); err != nil {
				return err
			}
		}
	}

	return nil
}

// fetchFileContent fetches the content of a file
func (c *HTTPClient) fetchFileContent(repo *Repository, filePath, downloadURL string) error {
	if downloadURL == "" {
		// Some files (like binary files) don't have a download URL
		repo.Files[filePath] = &File{
			Path: filePath,
			Type: "file",
		}
		return nil
	}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch file content: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	repo.Files[filePath] = &File{
		Path:    filePath,
		Content: string(content),
		Type:    "file",
	}

	return nil
}
