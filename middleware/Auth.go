package middleware

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/urlshortner/models"
)

func Auth(next http.Handler) http.Handler {
	secret := []byte(os.Getenv("JWT_SECRET")) // MUST be set in your env

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		idClaim, ok := claims["user_id"]
		if !ok {
			http.Error(w, "token missing user_id", http.StatusUnauthorized)
			return
		}

		var userID int

		switch v := idClaim.(type) {
		case string: // stored as "42"
			n, err := strconv.Atoi(v)
			if err != nil {
				http.Error(w, "bad user_id in token", http.StatusUnauthorized)
				return
			}
			userID = n

		case float64: // stored as 42 (JSON number → float64)
			userID = int(v)

		default:
			http.Error(w, "unknown user_id type", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), models.ContextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
