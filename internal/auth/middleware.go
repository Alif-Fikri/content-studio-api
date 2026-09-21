package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	keyfunc keyfunc.Keyfunc
}

func NewMiddleware(jwksURL string) (*Middleware, error) {
	kf, err := keyfunc.NewDefaultCtx(context.Background(), []string{jwksURL})
	if err != nil {
		return nil, err
	}
	return &Middleware{keyfunc: kf}, nil
}

const ContextUserIDKey = "user_id"

func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
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
