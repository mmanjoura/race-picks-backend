package users

import (
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

// Account godoc
// @Summary Get account details
// @Description Get account details
// @Tags users
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} object	"ok"
// @Router /account [get]
func Account(c *gin.Context) {

	// userId, err := strconv.Atoi(c.Param("id"))
	// Retrieve the value of the "jwt" cookie
	jwtCookie, err := c.Cookie("Authorization")
	if err != nil {
		// Handle the case when the cookie is not found or there's an error
		c.JSON(400, gin.H{"error": "Cookie not found"})
		return
	}

	token, err := jwt.ParseWithClaims(jwtCookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("k1U6pO+9qZteWy+yE52Z56qSBqmJ1orl27r/28AfkIA="), nil
	})

	if err != nil {
		// Handle parsing errors
		c.JSON(400, gin.H{"error": "Invalid token: " + err.Error()})
		return
	}

	if !token.Valid {
		// Handle invalid token
		c.JSON(400, gin.H{"error": "Invalid token"})
		return
	}

	claims := token.Claims.(*jwt.StandardClaims)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error while retreiving Accounts": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": claims})
}

func retrieveUserAccount(c *gin.Context, offset, limit int, userEmail string) (models.User, error) {
	db := database.Database.DB

	user := models.User{}

	err := db.QueryRowContext(c, `
		SELECT id,
		full_name,
		email,
		password,
		user_type,
		profile,
		avatar_url,
       Updated_At,
       Created_At
  FROM Users WHERE email = ?`, userEmail).
		Scan(&user.ID,
			&user.FullName,
			&user.Email,
			&user.Password,
			&user.UserType,
			&user.Profile,
			&user.AvatarUrl,
			&user.UpdatedAt,
			&user.CreatedAt,
		)

	if err != nil {
		return user, err
	}

	return user, nil
}
