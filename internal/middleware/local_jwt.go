package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LocalAuth struct {
	secret []byte
}

func NewLocalAuth(secret string) *LocalAuth {
	return &LocalAuth{secret: []byte(secret)}
}

func (a *LocalAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		// Dev token bypass
		if strings.HasPrefix(tokenStr, "dev-") {
			userID := tokenStr[4:]
			c.Set("user_id", userID)
			c.Set("username", "Dev_"+userID)
			c.Next()
			return
		}
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return a.secret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}
		if uid, ok := claims["user_id"].(string); ok {
			c.Set("user_id", uid)
		}
		if name, ok := claims["username"].(string); ok {
			c.Set("username", name)
		}
		c.Next()
	}
}
