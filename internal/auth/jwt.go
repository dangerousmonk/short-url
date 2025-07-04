package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/dangerousmonk/short-url/internal/logging"
)

const (
	tokenLifeTime = time.Hour * 24
	secretKeySize = 32
)

var (
	errExpiredToken  = errors.New("token: has expired")
	errInvalidToken  = errors.New("token: is invalid")
	errInvalidClaims = errors.New("claims: failed to initialize")
)

// Claims extends jwt.RegisteredClaims with userID
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

// Valid checks if token is expired
func (claims *Claims) Valid() error {
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return errExpiredToken
	}
	return nil
}

// JWTAuthenticator represents struct that implements
type JWTAuthenticator struct {
	secretKey string
}

// CreateToken generates new token for user, using golang-jwt package
func (maker *JWTAuthenticator) CreateToken(userID string, duration time.Duration) (string, error) {
	claims, err := NewClaims(userID, duration)
	if err != nil {
		logging.Log.Warnf("CreateToken NewClaims err %v", err)
		return "", errInvalidClaims
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(maker.secretKey))
}

// ValidateToken checks if provided token is valid
func (maker *JWTAuthenticator) ValidateToken(token string) (*Claims, error) {
	claims := &Claims{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			logging.Log.Warnf("ValidateToken wrong signing method: %v", token.Header["alg"])
			return nil, errInvalidToken
		}
		return []byte(maker.secretKey), nil
	}
	_, err := jwt.ParseWithClaims(token, claims, keyFunc)
	if err != nil {
		return nil, err
	}
	return claims, nil

}

// Authenticator describes interface that must be implemented for authorization middleware
type Authenticator interface {
	CreateToken(userID string, duration time.Duration) (string, error)
	ValidateToken(token string) (*Claims, error)
}

// NewJWTAuthenticator is a function that initialize Authenticator
func NewJWTAuthenticator(secretKey string) (Authenticator, error) {
	if len(secretKey) < secretKeySize {
		return nil, errors.New("invalid secretKey len")
	}
	return &JWTAuthenticator{secretKey}, nil
}

// NewJWTAuthenticator is a function that initialize jwt.Claims
func NewClaims(userID string, duration time.Duration) (*Claims, error) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
		UserID: userID,
	}
	return claims, nil
}
