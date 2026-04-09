package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type IAMAuth struct {
	jwksURL   string
	publicKey *rsa.PublicKey
	mu        sync.RWMutex
	lastFetch time.Time
}

func NewIAMAuth(jwksURL string) *IAMAuth {
	return &IAMAuth{jwksURL: jwksURL}
}

func (a *IAMAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		key, err := a.getPublicKey()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "auth service unavailable"})
			return
		}
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return key, nil
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

func (a *IAMAuth) getPublicKey() (*rsa.PublicKey, error) {
	a.mu.RLock()
	if a.publicKey != nil && time.Since(a.lastFetch) < time.Hour {
		defer a.mu.RUnlock()
		return a.publicKey, nil
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	resp, err := http.Get(a.jwksURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			N string `json:"n"`
			E string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}
	if len(jwks.Keys) == 0 {
		return nil, fmt.Errorf("no keys in JWKS")
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(jwks.Keys[0].N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwks.Keys[0].E)
	if err != nil {
		return nil, err
	}

	a.publicKey = &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}
	a.lastFetch = time.Now()
	return a.publicKey, nil
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return c.Query("token")
}

func GetUserID(c *gin.Context) string {
	if id, ok := c.Get("user_id"); ok {
		return id.(string)
	}
	return ""
}

func GetUsername(c *gin.Context) string {
	if name, ok := c.Get("username"); ok {
		return name.(string)
	}
	return ""
}
