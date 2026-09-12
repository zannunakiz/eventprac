package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequiredAuth() gin.HandlerFunc {
	return func(context *gin.Context) {

		// Check token exists
		tokenString := context.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Access denied, missing token!",
			})
			return
		}

		// Parse Token
		tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer"))
		token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// Validate token by extracting values
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			context.Set("userId", int(claims["sub"].(float64)))
			context.Next()
		} else {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
		}
	}
}
