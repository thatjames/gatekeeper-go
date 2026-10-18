package v1

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	jwtSecretVar = "GATEKEEPER_JWT_SECRET"
	tokenLife    = time.Hour * 24 * 7
)

var (
	jwtSecret []byte

	ErrNoJWTSecret = errors.New(jwtSecretVar + " must be set when Web is configured")
)

// LoadJWTSecret reads the signing secret from the environment. It fails when
// the variable is unset or empty, so the caller can refuse to start.
func LoadJWTSecret() error {
	secret := os.Getenv(jwtSecretVar)
	if secret == "" {
		return ErrNoJWTSecret
	}
	jwtSecret = []byte(secret)
	return nil
}

func CreateAuthToken(username string) (string, error) {
	if len(jwtSecret) == 0 {
		return "", ErrNoJWTSecret
	}
	claims := UserClaims{
		User: User{
			Username: username,
			Role:     UserRole,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenLife)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gatekeeper",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseAuthToken(tokenString string) (*UserClaims, error) {
	if len(jwtSecret) == 0 {
		return nil, ErrNoJWTSecret
	}
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
