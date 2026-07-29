package handler

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenLifetime = 8 * time.Hour
	tokenIssuer   = "todo-list"
)

type signInRequest struct {
	Password string `json:"password"`
}

type signInResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

type authClaims struct {
	PasswordHash string `json:"password_hash"`
	jwt.RegisteredClaims
}

func passwordHash(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func signingKey(password string) []byte {
	hash := sha256.Sum256(
		[]byte("todo-list-jwt:" + password),
	)

	return hash[:]
}

func passwordsEqual(first, second string) bool {
	firstHash := sha256.Sum256([]byte(first))
	secondHash := sha256.Sum256([]byte(second))

	return subtle.ConstantTimeCompare(firstHash[:], secondHash[:]) == 1
}

func createToken(password string) (string, error) {
	now := time.Now()

	claims := authClaims{
		PasswordHash: passwordHash(password),

		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifetime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(signingKey(password))
}

func validateToken(tokenString, password string) bool {
	if tokenString == "" {
		return false
	}

	claims := &authClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return signingKey(password), nil
		},
		jwt.WithValidMethods(
			[]string{jwt.SigningMethodHS256.Alg()},
		),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithIssuer(tokenIssuer),
	)
	if err != nil || !token.Valid {
		return false
	}

	expectedHash := passwordHash(password)

	return subtle.ConstantTimeCompare([]byte(claims.PasswordHash), []byte(expectedHash)) == 1
}

func SignIn(password string, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request signInRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, logger, http.StatusBadRequest, signInResponse{Error: "Некорректный JSON"})
			return
		}

		if !passwordsEqual(request.Password, password) {
			writeJSON(w, logger, http.StatusBadRequest, signInResponse{Error: "Неверный пароль"})
			return
		}

		token, err := createToken(password)
		if err != nil {
			writeJSON(w, logger, http.StatusInternalServerError, signInResponse{Error: fmt.Sprintf("Не удалось создать токен: %v", err)})
			return
		}

		writeJSON(w, logger, http.StatusOK, signInResponse{
			Token: token,
		})
	}
}

func RequireAuth(password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if password == "" {
					next.ServeHTTP(w, r)
					return
				}

				cookie, err := r.Cookie("token")
				if err != nil || !validateToken(cookie.Value, password) {
					http.Error(w, "Authentication required", http.StatusUnauthorized)
					return
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}
