package middleware

import (
	"net/http"
	"time"

	auth "github.com/mmanjoura/race-picks-backend/pkg/auth"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("Authorization")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "cannot Get Authorization token"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return auth.JwtKey, nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Cannot parse token"})
			c.Abort()
			return
		}

		claims := jwt.MapClaims{}
		var ok bool
		if claims, ok = token.Claims.(jwt.MapClaims); ok && token.Valid {
			exp := claims["exp"].(float64)
			if exp < float64(time.Now().Unix()) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
				c.Abort()
				return
			}

		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token when mapping Claims"})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		user, err := getUser(c, claims["iss"].(string))

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Error when getting user from token"})
			c.Abort()
			return
		}

		if user.ID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User is not found"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func getUser(c *gin.Context, userEmail string) (models.User, error) {
	db := database.Database.DB

	user := models.User{}

	err := db.QueryRowContext(c, `
	SELECT id,
       full_name,
       email,
       password,
	   user_type,
       Updated_At,
       Created_At
  FROM Users WHERE email = ?`, userEmail).
		Scan(&user.ID,
			&user.FullName,
			&user.Email,
			&user.Password,
			&user.UserType,
			&user.UpdatedAt,
			&user.CreatedAt,
		)

	if err != nil {
		return user, err
	}

	return user, nil
}
