package preparation

import (
	"net/http"

	"github.com/mmanjoura/race-picks-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

// Event represents an event with its name and time.
type Event struct {
	EventName string `json:"event_name"`
	EventTime string `json:"event_time"`
}

// / GetTodayMeeting godoc
// @Summary Get the today meeting
// @Description Get the today meeting
// @Tags GetTodayMeeting
// @Accept  json
// @Produce  json
// @Success 200 {object} []Event
// @Router /analytics/today-meeting [get]
func GetTodayMeeting(c *gin.Context) {

	db := database.Database.DB

	eventDate := c.Query("date")
	eventName := c.Query("event_name")

	// Execute the query
	rows, err := db.Query(`
				SELECT 	event_name, 
						GROUP_CONCAT(event_time ORDER BY event_time) AS event_times 
				FROM EventRunners WHERE DATE(event_date) = ? and event_name in (SELECT event_name FROM events where country = ?) GROUP BY event_name ORDER BY event_name;`, eventDate, eventName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Create a slice to hold the events
	var events []Event

	// Loop through the rows and append the results to the slice
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.EventName, &event.EventTime); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		events = append(events, event)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the events
	c.JSON(http.StatusOK, events)
}
