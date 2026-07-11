package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	RoleKey     contextKey = "role"     // Переименовали в RoleKey, так как теперь храним строку
	UsernameKey contextKey = "username"
)

func AuthMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "missing authorization header"}`))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "invalid authorization format"}`))
				return
			}

			tokenString := parts[1]

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return jwtSecret, nil
			})

			if err != nil || !token.Valid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "invalid or expired token"}`))
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "invalid token claims"}`))
				return
			}

			ctx := r.Context()
			if userID, ok := claims["user_id"].(float64); ok {
				ctx = context.WithValue(ctx, UserIDKey, int64(userID))
			}
			
			// Извлекаем "role" как строку ("reader", "editor", "admin") вместо числа
			if role, ok := claims["role"].(string); ok {
				ctx = context.WithValue(ctx, RoleKey, role)
			}
			
			if username, ok := claims["username"].(string); ok {
				ctx = context.WithValue(ctx, UsernameKey, username)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}