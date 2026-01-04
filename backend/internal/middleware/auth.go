package middleware

import (
	"net/http"
	"strings"

	"github.com/expensesplit/backend/internal/appcontext"
	"github.com/expensesplit/backend/internal/services"
	"github.com/expensesplit/backend/pkg/utils"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.Unauthorized(w, "Authorization header is required")
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Unauthorized(w, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			utils.Unauthorized(w, "Invalid or expired token")
			return
		}

		// Add user info to context
		ctx := appcontext.WithUserID(r.Context(), claims.UserID)
		ctx = appcontext.WithUserEmail(ctx, claims.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
