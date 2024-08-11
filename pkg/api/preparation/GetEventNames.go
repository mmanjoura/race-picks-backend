package preparation

import (
	"net/http"

	"github.com/mmanjoura/race-picks-backend/pkg/database"

	"github.com/gin-gonic/gin"
)


func GetEventNames(c *gin.Context) {
	db := database.Database.DB


	// Get distinct Event name from SelectionsForm table
	rows, err := db.Query(`
		SELECT DISTINCT racecourse FROM SelectionsForm
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var eventNames []string	
	for rows.Next() {
		var eventName string
		err := rows.Scan(
			&eventName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		eventNames = append(eventNames, eventName)
	}	

	c.JSON(http.StatusOK, gin.H{"event names": eventNames})
}

