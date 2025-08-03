package server

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestGenerateJWT(t *testing.T) {
	t.Run("generates valid JWT with user_id", func(t *testing.T) {
		tokenStr, err := generateJWT(123)
		require.NoError(t, err)
		require.NotEmpty(t, tokenStr)

		uid, err := parseJWT(tokenStr)
		require.NoError(t, err)
		require.Equal(t, uint64(123), uid)
	})
}

func TestParseJWT(t *testing.T) {
	t.Run("returns user_id from valid token", func(t *testing.T) {
		tokenStr, _ := generateJWT(555)

		uid, err := parseJWT(tokenStr)
		require.NoError(t, err)
		require.Equal(t, uint64(555), uid)
	})

	t.Run("fails on tampered token", func(t *testing.T) {
		tokenStr, _ := generateJWT(42)

		// Подделываем подпись (например, меняем символ)
		tampered := tokenStr[:len(tokenStr)-1] + "x"

		uid, err := parseJWT(tampered)
		require.Error(t, err)
		require.Equal(t, uint64(0), uid)
		require.Contains(t, err.Error(), "невалидный токен")
	})

	t.Run("fails on invalid structure", func(t *testing.T) {
		uid, err := parseJWT("not-a-token")
		require.Error(t, err)
		require.Equal(t, uint64(0), uid)
		require.Contains(t, err.Error(), "невалидный токен")
	})

	t.Run("fails when claims are not map", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		})

		tokenStr, _ := token.SignedString(jwtSecret)

		uid, err := parseJWT(tokenStr)
		require.Error(t, err)
		require.Equal(t, uint64(0), uid)
		require.Contains(t, err.Error(), "user_id отсутствует")
	})

	t.Run("fails when user_id missing", func(t *testing.T) {
		claims := jwt.MapClaims{
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := token.SignedString(jwtSecret)

		uid, err := parseJWT(tokenStr)
		require.Error(t, err)
		require.Equal(t, uint64(0), uid)
		require.Contains(t, err.Error(), "user_id отсутствует")
	})

	t.Run("fails when user_id wrong type", func(t *testing.T) {
		claims := jwt.MapClaims{
			"user_id": "not-a-number",
			"exp":     time.Now().Add(time.Hour).Unix(),
			"iat":     time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := token.SignedString(jwtSecret)

		uid, err := parseJWT(tokenStr)
		require.Error(t, err)
		require.Equal(t, uint64(0), uid)
		require.Contains(t, err.Error(), "user_id отсутствует")
	})

	t.Run("fails on expired token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"user_id": float64(777),
			"exp":     time.Now().Add(-time.Minute).Unix(), // уже истёк
			"iat":     time.Now().Add(-2 * time.Minute).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := token.SignedString(jwtSecret)

		uid, err := parseJWT(tokenStr)
		require.Error(t, err)
		require.Equal(t, uint64(0), uid)
		require.Contains(t, err.Error(), "невалидный токен")
	})
}
