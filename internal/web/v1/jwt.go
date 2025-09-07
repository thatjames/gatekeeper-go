package v1

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateAuthToken(username string) (string, error) {
	claims := UserClaims{
		User: User{
			Username: username,
			Role:     UserRole,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gatekeeper",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("secret"))
}

func ParseAuthToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
