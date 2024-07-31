package preparation

import (
	"net/http"

	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

// / GetMeetingRunners godoc
// @Summary Get the meeting runners
// @Description Get the meeting runners
// @Tags GetMeetingRunners
// @Accept  json
// @Produce  json
// @Param event_name query string true "Event Name"
// @Success 200 {object} models.TodayRunners
// @Router /analytics/meeting-runners [get]
func GetMeetingRunners(c *gin.Context) {

	db := database.Database.DB
	eventName := c.Query("event_name")

	// Execute the query
	rows, err := db.Query("SELECT id, selection_name, event_time, event_name, price, selection_id FROM TodayRunners WHERE DATE(event_date) = DATE('now') AND event_name = ?", eventName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Create a slice to hold the meeting data
	var meetingsData []models.TodayRunners

	// Loop through the rows and append the results to the slice
	for rows.Next() {
		var meetingData models.TodayRunners
		err := rows.Scan(
			&meetingData.ID,
			&meetingData.SelectionName,
			&meetingData.EventTime,
			&meetingData.EventName,
			&meetingData.Price,
			&meetingData.SelectionID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		meetingsData = append(meetingsData, meetingData)
	}

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the meeting data
	c.JSON(http.StatusOK, gin.H{"meetingData": meetingsData})
}
