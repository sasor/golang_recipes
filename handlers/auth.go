package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sasor/golang_recipes/envs"
	"github.com/sasor/golang_recipes/models"
	"net/http"
	"time"
)

type AuthHandler struct{}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (a *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if user.Username != "root" && user.Password != "password" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
		return
	}

	// https://pkg.go.dev/github.com/golang-jwt/jwt/v4#NewWithClaims
	expTime := time.Now().Add(time.Minute * 10)
	claims := &Claims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			Issuer:    "test",
		},
	}

	jwtSecret := envs.Get("JWT_SECRET", "default")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}

func (a *AuthHandler) RefreshToken(c *gin.Context) {
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

	t := time.Unix(claims.ExpiresAt.Unix(), 0).Sub(time.Now())
	fmt.Printf("NOW:::::: %v\n", t)
	u := time.Second * 30
	fmt.Printf("ANOTHER::::: %v\n", u)

	if t > u {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token is not expired yet",
		})
		return
	}

	newExpirationTime := time.Now().Add(time.Minute * 10)
	claims.ExpiresAt = jwt.NewNumericDate(newExpirationTime)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshTokenValue, err := refreshToken.SignedString([]byte(envs.Get("JWT_SECRET", "default")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": refreshTokenValue,
	})
}
