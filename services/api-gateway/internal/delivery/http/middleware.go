package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/romanlovesweed/yammi/services/api-gateway/internal/infrastructure"
)

type contextKey string

const userIDKey contextKey = "user_id"

// UserIDFromContext извлекает user_id из контекста (установленный AuthMiddleware).
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

// AuthMiddleware проверяет JWT из заголовка Authorization: Bearer <token>.
// При успехе кладёт user_id в контекст.
func AuthMiddleware(verifier *infrastructure.JWTVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				writeError(w, http.StatusUnauthorized, "authorization token required")
				return
			}

			if !strings.HasPrefix(header, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "invalid authorization format, expected: Bearer <token>")
				return
			}

			tokenString := strings.TrimPrefix(header, "Bearer ")

			userID, err := verifier.VerifyToken(tokenString)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OwnerOnly проверяет что user_id из токена совпадает с {id} из URL.
// Должен применяться ПОСЛЕ AuthMiddleware.
func OwnerOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenUserID, ok := UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "authorization required")
			return
		}

		pathID := r.PathValue("id")
		if pathID != tokenUserID {
			writeError(w, http.StatusForbidden, "access denied: you can only manage your own account")
			return
		}

		next(w, r)
	}
}
