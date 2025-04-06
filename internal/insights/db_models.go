package insights

import (
	"encoding/json"
	"time"
)

// Base insight record with common fields
type BaseInsightRecord struct {
	ID           int64     `db:"id"`
	RepositoryID int64     `db:"repository_id"`
	Model        string    `db:"model"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// FunctionInsightRecord represents a stored function insight in the database
type FunctionInsightRecord struct {
	BaseInsightRecord
	FunctionID int64  `db:"function_id"`
	Data       []byte `db:"data"` // JSONB data
}

// Unmarshals the JSONB data into a FunctionInsight
func (r *FunctionInsightRecord) GetInsight() (*FunctionInsight, error) {
	insight := &FunctionInsight{}
	err := json.Unmarshal(r.Data, insight)
	return insight, err
}

// SymbolInsightRecord represents a stored symbol insight in the database
type SymbolInsightRecord struct {
	BaseInsightRecord
	SymbolID int64  `db:"symbol_id"`
	Data     []byte `db:"data"` // JSONB data
}

// Unmarshals the JSONB data into a SymbolInsight
func (r *SymbolInsightRecord) GetInsight() (*SymbolInsight, error) {
	insight := &SymbolInsight{}
	err := json.Unmarshal(r.Data, insight)
	return insight, err
}

// StructInsightRecord represents a stored struct insight in the database
type StructInsightRecord struct {
	BaseInsightRecord
	SymbolID int64  `db:"symbol_id"`
	Data     []byte `db:"data"` // JSONB data
}

// Unmarshals the JSONB data into a StructInsight
func (r *StructInsightRecord) GetInsight() (*StructInsight, error) {
	insight := &StructInsight{}
	err := json.Unmarshal(r.Data, insight)
	return insight, err
}

// FileInsightRecord represents a stored file insight in the database
type FileInsightRecord struct {
	BaseInsightRecord
	FileID int64  `db:"file_id"`
	Data   []byte `db:"data"` // JSONB data
}

// Unmarshals the JSONB data into a FileInsight
func (r *FileInsightRecord) GetInsight() (*FileInsight, error) {
	insight := &FileInsight{}
	err := json.Unmarshal(r.Data, insight)
	return insight, err
}

// RepositoryInsightRecord represents a stored repository insight in the database
type RepositoryInsightRecord struct {
	BaseInsightRecord
	Data []byte `db:"data"` // JSONB data
}

// Unmarshals the JSONB data into a RepositoryInsight
func (r *RepositoryInsightRecord) GetInsight() (*RepositoryInsight, error) {
	insight := &RepositoryInsight{}
	err := json.Unmarshal(r.Data, insight)
	return insight, err
}
