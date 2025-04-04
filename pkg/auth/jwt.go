package auth

import (
	"errors"
	"fmt"
	"time"

	"cred.com/hack25/backend/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType defines the type of token
type TokenType string

const (
	// AccessToken is used for API access
	AccessToken TokenType = "access"
	// RefreshToken is used to generate new access tokens
	RefreshToken TokenType = "refresh"
)

// JWTClaims represents the claims in the JWT
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Type   TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTService is responsible for JWT operations
type JWTService struct {
	secretKey       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	signingMethod   jwt.SigningMethod
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, accessTokenTTL, refreshTokenTTL time.Duration, signingAlg string) *JWTService {
	var signingMethod jwt.SigningMethod
	switch signingAlg {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "HS384":
		signingMethod = jwt.SigningMethodHS384
	case "HS512":
		signingMethod = jwt.SigningMethodHS512
	default:
		signingMethod = jwt.SigningMethodHS256
	}

	return &JWTService{
		secretKey:       secretKey,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		signingMethod:   signingMethod,
	}
}

// GenerateAccessToken generates a new access token
func (s *JWTService) GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
	return s.generateToken(userID, email, role, AccessToken, s.accessTokenTTL)
}

// GenerateRefreshToken generates a new refresh token
func (s *JWTService) GenerateRefreshToken(userID uuid.UUID, email, role string) (string, error) {
	return s.generateToken(userID, email, role, RefreshToken, s.refreshTokenTTL)
}

// generateToken generates a new token
func (s *JWTService) generateToken(userID uuid.UUID, email, role string, tokenType TokenType, expiration time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(s.signingMethod, claims)
	signedToken, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		logger.Errorf("Failed to generate token: %v", err)
		return "", err
	}

	return signedToken, nil
}

// ValidateToken validates and parses a token
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		logger.Errorf("Failed to parse token: %v", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// TokenResponse represents the token data to be returned in API responses
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Seconds until the access token expires
}

// GenerateTokenPair generates an access and refresh token pair
func (s *JWTService) GenerateTokenPair(userID uuid.UUID, email, role string) (*TokenResponse, error) {
	accessToken, err := s.GenerateAccessToken(userID, email, role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateRefreshToken(userID, email, role)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
	}, nil
}
