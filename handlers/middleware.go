package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sasor/golang_recipes/envs"
	"net/http"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X_API_KEY") != envs.Get("X_API_KEY", "") {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		c.Next()
	}
}
