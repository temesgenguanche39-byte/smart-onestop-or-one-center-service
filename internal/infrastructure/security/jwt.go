package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/domain"
)

type JWTClaims struct {
	UserID      uuid.UUID       `json:"sub"`
	Email       string          `json:"email"`
	Role        domain.UserRole `json:"role"`
	StructureID *uint           `json:"structure_id,omitempty"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

func (m *JWTManager) GenerateToken(user *domain.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(m.tokenDuration)
	claims := &JWTClaims{
		UserID:      user.ID,
		Email:       user.Email,
		Role:        user.Role,
		StructureID: user.StructureID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "smart-onestop-platform",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expiresAt, nil
}

func (m *JWTManager) Verify(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
