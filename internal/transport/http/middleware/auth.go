package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	jwtutil "github.com/horoshi10v/tires-shop/pkg/jwt"
)

// RequireRole checks the JWT token and ensures the user has one of the allowed roles.
func RequireRole(jwtSecret string, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format, expected 'Bearer <token>'"})
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 2. Parse and validate the JWT token
		token, err := jwt.ParseWithClaims(tokenString, &jwtutil.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// 3. Extract claims (UserID and Role)
		claims, ok := token.Claims.(*jwtutil.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		// 4. Check if the role is allowed
		isAllowed := false
		for _, role := range allowedRoles {
			if claims.Role == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden, insufficient permissions"})
			c.Abort()
			return
		}

		// 5. Store UserID and Role in context for Handlers/Services to use
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}
