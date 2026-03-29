package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	jwtutil "github.com/horoshi10v/tires-shop/pkg/jwt"
)

// OptionalAuth parses a Bearer token when present and stores user claims in context.
// Invalid or missing tokens are ignored so public endpoints can still proceed as guests.
func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token, err := jwt.ParseWithClaims(parts[1], &jwtutil.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.Next()
			return
		}

		claims, ok := token.Claims.(*jwtutil.Claims)
		if !ok {
			c.Next()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}
