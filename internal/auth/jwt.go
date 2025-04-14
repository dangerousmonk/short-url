package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenLifeTime = time.Hour * 3
	secretKeySize = 32
)

var (
	errExpiredToken = errors.New("token has expired")
	errInvalidToken = errors.New("token is invalid")
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

type JWTAuthenticator struct {
	secretKey string
}

func (claims *Claims) Valid() error {
	if time.Now().After(claims.ExpiresAt.Time) {
		return errExpiredToken
	}
	return nil
}

type Authenticator interface {
	CreateToken(userID string, duration time.Duration) (string, error)
	ValidateToken(token string) (*Claims, error)
}

func (maker *JWTAuthenticator) CreateToken(userID string, duration time.Duration) (string, error) {
	claims, err := NewClaims(userID, duration)
	if err != nil {
		return "", nil
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(maker.secretKey))
}
func (maker *JWTAuthenticator) ValidateToken(token string) (*Claims, error) {
	claims := &Claims{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errInvalidToken
		}
		return []byte(maker.secretKey), nil
	}
	jwtToken, err := jwt.ParseWithClaims(token, claims, keyFunc)
	if err != nil {
		return nil, err
	}
	if !jwtToken.Valid {
		return nil, errInvalidToken
	}
	return claims, nil

}

func NewJWTAuthenticator(secretKey string) (Authenticator, error) {
	if len(secretKey) < secretKeySize {
		return nil, errors.New("invalid secretKey len")
	}
	return &JWTAuthenticator{secretKey}, nil
}

func NewClaims(userID string, duration time.Duration) (*Claims, error) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
		UserID: userID,
	}
	return claims, nil
}
