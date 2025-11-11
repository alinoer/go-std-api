package service

import (
	"fmt"
	"time"

	"github.com/alinoer/go-std-api/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService struct {
	config models.TokenConfig
}

func NewAuthService(secretKey string) *AuthService {
	return &AuthService{
		config: models.TokenConfig{
			SecretKey: secretKey,
			ExpiresIn: 24 * time.Hour, // 24 hours
			Issuer:    "go-std-api",
		},
	}
}

func (s *AuthService) GenerateToken(userID uuid.UUID, username string) (string, int64, error) {
	now := time.Now()
	expirationTime := now.Add(s.config.ExpiresIn)

	claims := models.JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, int64(s.config.ExpiresIn.Seconds()), nil
}

func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (s *AuthService) RefreshToken(oldTokenString string) (string, int64, error) {
	claims, err := s.ValidateToken(oldTokenString)
	if err != nil {
		return "", 0, fmt.Errorf("invalid token for refresh: %w", err)
	}

	// Check if token is close to expiration (within 1 hour)
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", 0, fmt.Errorf("token is still valid, refresh not needed")
	}

	return s.GenerateToken(claims.UserID, claims.Username)
}