package preparation

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

func GetRacingMarketWinners(c *gin.Context) {
	db := database.Database.DB

	currentTime := time.Now()
	// Subtract one day to get the day before
	dayBefore := currentTime.AddDate(0, 0, -1)
	// Format the date as YYYY-MM-DD
	formattedDate := dayBefore.Format("2006-01-02")
	rows, err := db.Query(`select selection_id, selection_link, selection_name from EventRunners WHERE  DATE(event_date) = ?`, formattedDate)
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
		err = SaveSelectionsForm(db, c, selection.ID, selection.Link, selection.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "postion updated successfully"})
}
