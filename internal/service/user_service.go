package service

import (
	"errors"

	"cred.com/hack25/backend/internal/models"
	"cred.com/hack25/backend/internal/repository"
	"cred.com/hack25/backend/pkg/auth"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/google/uuid"
)

// UserService handles user business logic
type UserService struct {
	userRepo  *repository.UserRepository
	jwtService *auth.JWTService
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository, jwtService *auth.JWTService) *UserService {
	return &UserService{
		userRepo:  userRepo,
		jwtService: jwtService,
	}
}

// RegisterUser registers a new user
func (s *UserService) RegisterUser(user *models.User) (*models.UserResponse, error) {
	// Check if email already exists
	exists, err := s.userRepo.EmailExists(user.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		logger.Warnf("Email already exists: %s", user.Email)
		return nil, errors.New("email already exists")
	}

	// Create user
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// AuthenticateUser authenticates a user and returns tokens
func (s *UserService) AuthenticateUser(email, password string) (*auth.TokenResponse, *models.UserResponse, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		logger.Warnf("Authentication failed: %v", err)
		return nil, nil, errors.New("invalid credentials")
	}

	if !user.CheckPassword(password) {
		logger.Warnf("Invalid password for user: %s", email)
		return nil, nil, errors.New("invalid credentials")
	}

	// Generate JWT tokens
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, nil, err
	}

	response := user.ToResponse()
	return tokens, &response, nil
}

// GetUserByID gets a user by ID
func (s *UserService) GetUserByID(id uuid.UUID) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(id uuid.UUID, firstName, lastName string) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	user.FirstName = firstName
	user.LastName = lastName

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *UserService) RefreshToken(refreshToken string) (*auth.TokenResponse, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		logger.Warnf("Invalid refresh token: %v", err)
		return nil, errors.New("invalid refresh token")
	}

	if claims.Type != auth.RefreshToken {
		logger.Warnf("Token is not a refresh token")
		return nil, errors.New("invalid token type")
	}

	// Get user
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// ListUsers lists all users with pagination
func (s *UserService) ListUsers(page, pageSize int) ([]models.UserResponse, int64, error) {
	users, total, err := s.userRepo.List(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, total, nil
}
