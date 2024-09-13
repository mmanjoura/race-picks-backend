package preparation

import (
	"net/http"


	"github.com/gin-gonic/gin"

	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

func GetRacingMarketWinners(c *gin.Context) {
	db := database.Database.DB

	var raceDate models.EventDate



	// Bind JSON input to optimalParams
	if err := c.ShouldBindJSON(&raceDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}


	rows, err := db.Query(`
							select 
								selection_id, 
								selection_link, 
								selection_name 
							from EventRunners 								
								where DATE(event_date) = ?`, raceDate.Date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	selections := []models.Selection{}
	for rows.Next() {
		var selection models.Selection
		err := rows.Scan(&selection.ID, &selection.Link, &selection.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		selections = append(selections, selection)
	}
	defer rows.Close()

	for _, selection := range selections {
		err = SaveSelectionsForm(db, c, selection.ID, selection.Link, selection.Name, true, raceDate.Date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "postion updated successfully"})
}
