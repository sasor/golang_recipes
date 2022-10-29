package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sasor/golang_recipes/envs"
	"net/http"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenFromHeader := c.GetHeader("Authorization")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(
			tokenFromHeader,
			claims,
			func(token *jwt.Token) (interface{}, error) {
				return []byte(envs.Get("JWT_SECRET", "default")), nil
			},
		)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		if token == nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		c.Next()
	}
}
