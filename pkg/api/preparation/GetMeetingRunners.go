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
func GetMeetingRunners(c *gin.Context) {	db := database.Database.DB
	type Selection struct {
		ID   int
		Name string
	}
	eventName := c.Query("event_name")
	eventTime := c.Query("event_time")

	// Get today's runners for the given event_name and event_date
	rows, err := db.Query("SELECT selection_id FROM TodayRunners WHERE  event_name = ? and event_time = ?", eventName, eventTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var analysisData []models.AnalysisData
	var selections []Selection
	for rows.Next() {
		var selection Selection
		err := rows.Scan(&selection.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if selection.ID == 0 {
			continue
		}
		selections = append(selections, selection)
	}

	for _, selection := range selections {

		// Execute the query
		rows, err = db.Query(`
						SELECT
							selection_id,
							selection_name,							
							COUNT(*) AS num_runs,
							AVG(CAST(SUBSTR(position, 1, INSTR(position, '/') - 1) AS INTEGER)) AS avg_position,
							AVG(rating) AS avg_rating,
							AVG(
								CASE
									WHEN distance LIKE '%y' THEN
										-- Convert distance to furlongs
										CAST(SUBSTR(distance, 1, INSTR(distance, 'm') - 1) AS INTEGER) * 8
										+ CAST(SUBSTR(distance, INSTR(distance, 'm') + 1, INSTR(distance, 'f') - INSTR(distance, 'm') - 1) AS INTEGER)
										+ CAST(REPLACE(SUBSTR(distance, INSTR(distance, 'f') + 1), 'y', '') AS INTEGER) / 220.0
									WHEN distance LIKE '%f' THEN
										-- Distance is already in furlongs
										CAST(REPLACE(distance, 'f', '') AS INTEGER)
									ELSE
										-- Convert distance in meters to furlongs
										CAST(REPLACE(distance, 'm', '') AS INTEGER) / 201.168
								END
							) AS avg_distance_furlongs,
							AVG(
								CAST(SUBSTR(sp_odds, 1, INSTR(sp_odds, '/') - 1) AS FLOAT) / 
								CAST(SUBSTR(sp_odds, INSTR(sp_odds, '/') + 1) AS FLOAT) + 1
							) AS avg_odds
						FROM
							SelectionsForm	WHERE selection_id = ? order by num_runs asc`, selection.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		

		// Loop through the rows and append the results to the slice
		var data models.AnalysisData
		for rows.Next() {		
			err := rows.Scan(
				&data.SelectionID,
				&data.SelectionName,
				&data.NumRuns,
				&data.AvgPosition,
				&data.AvgRating,
				&data.AvgDistanceFurlongs,
				&data.AvgOdds,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			
		}
		analysisData = append(analysisData, data)
	}

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the meeting data
c.JSON(http.StatusOK, gin.H{"meetingData": analysisData})
}
