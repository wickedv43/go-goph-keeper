package server

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

// jwtSecret is the secret key used to sign JWT tokens.
var jwtSecret = []byte("my-very-secret-key")

// generateJWT generates a JWT token with the given user ID and 24-hour expiration.
func generateJWT(userID uint64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// parseJWT validates the token and extracts the user ID from its claims.
func parseJWT(tokenStr string) (uint64, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("невалидный токен")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("невалидный payload")
	}

	uidFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("user_id отсутствует или некорректен")
	}

	return uint64(uidFloat), nil
}
