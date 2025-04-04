package repository

import (
	"database/sql"
	"errors"
	"time"

	"cred.com/hack25/backend/internal/models"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/google/uuid"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	logger.Infof("Creating user with email: %s", user.Email)

	if err := user.HashPassword(); err != nil {
		logger.Errorf("Failed to hash password: %v", err)
		return err
	}

	query := `
	INSERT INTO users (id, email, password, first_name, last_name, active, role, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.Active,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		logger.Errorf("Failed to create user: %v", err)
		return err
	}

	logger.Infof("User created successfully with ID: %s", user.ID)
	return nil
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	query := `
	SELECT id, email, password, first_name, last_name, active, role, created_at, updated_at, deleted_at
	FROM users
	WHERE id = $1 AND deleted_at IS NULL
	`

	row := r.db.QueryRow(query, id)

	user := &models.User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Active,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warnf("User not found with ID: %s", id)
			return nil, errors.New("user not found")
		}
		logger.Errorf("Failed to get user by ID: %v", err)
		return nil, err
	}

	return user, nil
}

// GetByEmail gets a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
	SELECT id, email, password, first_name, last_name, active, role, created_at, updated_at, deleted_at
	FROM users
	WHERE email = $1 AND deleted_at IS NULL
	`

	row := r.db.QueryRow(query, email)

	user := &models.User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Active,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warnf("User not found with email: %s", email)
			return nil, errors.New("user not found")
		}
		logger.Errorf("Failed to get user by email: %v", err)
		return nil, err
	}

	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	logger.Infof("Updating user with ID: %s", user.ID)

	query := `
	UPDATE users
	SET first_name = $1, last_name = $2, active = $3, role = $4, updated_at = $5
	WHERE id = $6 AND deleted_at IS NULL
	`

	user.UpdatedAt = time.Now()

	res, err := r.db.Exec(
		query,
		user.FirstName,
		user.LastName,
		user.Active,
		user.Role,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		logger.Errorf("Failed to update user: %v", err)
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		logger.Errorf("Failed to get affected rows: %v", err)
		return err
	}

	if affected == 0 {
		logger.Warnf("User not found with ID: %s", user.ID)
		return errors.New("user not found")
	}

	logger.Infof("User updated successfully")
	return nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id uuid.UUID) error {
	logger.Infof("Deleting user with ID: %s", id)

	query := `
	UPDATE users
	SET deleted_at = $1
	WHERE id = $2 AND deleted_at IS NULL
	`

	res, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		logger.Errorf("Failed to delete user: %v", err)
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		logger.Errorf("Failed to get affected rows: %v", err)
		return err
	}

	if affected == 0 {
		logger.Warnf("User not found with ID: %s", id)
		return errors.New("user not found")
	}

	logger.Infof("User deleted successfully")
	return nil
}

// List lists all users with pagination
func (r *UserRepository) List(page, pageSize int) ([]models.User, int64, error) {
	// Count total users
	countQuery := `
	SELECT COUNT(*) FROM users WHERE deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(countQuery).Scan(&count)
	if err != nil {
		logger.Errorf("Failed to count users: %v", err)
		return nil, 0, err
	}

	// Get paginated users
	offset := (page - 1) * pageSize
	query := `
	SELECT id, email, password, first_name, last_name, active, role, created_at, updated_at, deleted_at
	FROM users
	WHERE deleted_at IS NULL
	ORDER BY created_at DESC
	LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, pageSize, offset)
	if err != nil {
		logger.Errorf("Failed to list users: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.Active,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			logger.Errorf("Failed to scan user: %v", err)
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, count, nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(email string) (bool, error) {
	query := `
	SELECT COUNT(*) FROM users WHERE email = $1 AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(query, email).Scan(&count)
	if err != nil {
		logger.Errorf("Failed to check if email exists: %v", err)
		return false, err
	}

	return count > 0, nil
}
