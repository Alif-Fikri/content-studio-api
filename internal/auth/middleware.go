package auth

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	keyfunc keyfunc.Keyfunc
	apiKey  string
}

func NewMiddleware(jwksURL, apiKey string) (*Middleware, error) {
	kf, err := keyfunc.NewDefaultCtx(context.Background(), []string{jwksURL})
	if err != nil {
		return nil, err
	}
	return &Middleware{keyfunc: kf, apiKey: apiKey}, nil
}

const ContextUserIDKey = "user_id"

func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.apiKey != "" {
			provided := c.GetHeader("X-API-Key")
			if provided != "" {
				if subtle.ConstantTimeCompare([]byte(provided), []byte(m.apiKey)) == 1 {
					c.Set(ContextUserIDKey, "service")
					c.Next()
					return
				}
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
				return
			}
		}

		header := c.GetHeader("Authorization")
		token, found := strings.CutPrefix(header, "Bearer ")
		if !found || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		parsed, err := jwt.Parse(token, m.keyfunc.Keyfunc)
		if err != nil || !parsed.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		sub, _ := claims["sub"].(string)
		c.Set(ContextUserIDKey, sub)
		c.Next()
	}
}
