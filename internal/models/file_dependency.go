package models

import (
	"time"
)

// FileDependency represents a file-level import dependency
type FileDependency struct {
	ID           int64     `json:"id" db:"id"`
	RepositoryID int64     `json:"repository_id" db:"repository_id"`
	FileID       int64     `json:"file_id" db:"file_id"`
	ImportPath   string    `json:"import_path" db:"import_path"`
	Alias        string    `json:"alias,omitempty" db:"alias"`
	IsStdlib     bool      `json:"is_stdlib" db:"is_stdlib"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
