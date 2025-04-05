package database

import (
	"database/sql"
	"fmt"
	"time"

	"cred.com/hack25/backend/pkg/logger"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Config holds the database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DB is a wrapper for the SQL DB
type DB struct {
	Conn *sql.DB
}

// NewDB creates a new PostgreSQL database connection
func NewDB(config Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	logger.Infof("Connecting to PostgreSQL database at %s:%s", config.Host, config.Port)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Errorf("Failed to connect to database: %v", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		logger.Errorf("Failed to ping database: %v, dsn: %s", err, dsn)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	logger.Info("Successfully connected to PostgreSQL database")

	return &DB{Conn: db}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.Conn.Close()
}

// InitSchema initializes the database schema
func (d *DB) InitSchema() error {
	logger.Info("Initializing database schema")

	// Create users table
	userTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		first_name VARCHAR(100),
		last_name VARCHAR(100),
		active BOOLEAN DEFAULT TRUE,
		role VARCHAR(50) DEFAULT 'user',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE
	);
	`

	_, err := d.Conn.Exec(userTableSQL)
	if err != nil {
		logger.Errorf("Failed to create users table: %v", err)
		return fmt.Errorf("failed to create users table: %w", err)
	}

	logger.Info("Database schema initialized successfully")
	return nil
}

// Transaction executes a function within a database transaction
func (d *DB) Transaction(fn func(*sql.Tx) error) error {
	tx, err := d.Conn.Begin()
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
