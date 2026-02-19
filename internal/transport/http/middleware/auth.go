package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequireRole checks if the user has the required role to access the endpoint.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get user role from headers (Simulating JWT claims)
		userRole := c.GetHeader("X-User-Role")

		if userRole == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized, missing role header"})
			c.Abort() // Stop the request chain
			return
		}

		// 2. Check if the user's role is in the list of allowed roles
		isAllowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				isAllowed = true
				break
			}
		}

		// 3. If not allowed, return 403 Forbidden
		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden, insufficient permissions"})
			c.Abort()
			return
		}

		// --- НОВОЕ: Имитируем ID авторизованного сотрудника ---
		// В реальном проекте этот ID берется из JWT токена (claims.UserID)
		dummyAdminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		c.Set("userID", dummyAdminID)
		// --------------------------------------------------------

		c.Next()
	}
}
