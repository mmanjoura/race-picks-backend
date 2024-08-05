package analysis

// import (
// 	"fmt"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"github.com/mmanjoura/race-picks-backend/pkg/database"
// 	"github.com/mmanjoura/race-picks-backend/pkg/models"
// )

// func GetAverages_3(c *gin.Context) {
// 	db := database.Database.DB
// 	var modelparams models.AnalysisData

// 	if err := c.ShouldBindJSON(&modelparams); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Get today's runners for the given event_name and event_date
// 	rows, err := db.Query("SELECT selection_id, selection_name FROM TodayRunners WHERE event_name = ? AND DATE(event_date) = DATE('now') AND event_time = ?", modelparams.EventName, modelparams.EventTime)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	// Create a slice to hold the selections
// 	var selections []Selection
// 	for rows.Next() {
// 		var selection Selection
// 		err := rows.Scan(&selection.ID, &selection.Name)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		selections = append(selections, selection)
// 	}

// 	// Calculate the average of the BSP and other fields
// 	var eventResults []models.AnalysisData
// 	for _, selection := range selections {
// 		var eventResult models.AnalysisData

// 	fmt.Println("selection.Name: ", selection.Name)
	

// 		if err != nil {
// 			if err.Error() == "sql: no rows in result set" {
// 				continue
// 			}
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}



// 			eventResults = append(eventResults, eventResult)

// 		}

	


// 	c.JSON(http.StatusOK, gin.H{"message": "Horse information saved successfully", "data": eventResults})
// }

