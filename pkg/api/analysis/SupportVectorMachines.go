// Explanation
// Data Preparation:

// The handler retrieves selections from the database based on the provided event_name and event_time.
// It formats this data into a JSON structure suitable for sending to the external SVM service.
// External SVM Service:

// The handler sends a POST request with the formatted data to an external SVM service running on localhost:5000/predict_svm.
// Adjust the URL as necessary depending on where your SVM model is hosted.
// Processing Predictions:

// After receiving the predictions from the SVM service, the handler maps these probabilities to the corresponding selections.
// It uses a threshold (e.g., 0.5) to determine if a selection is predicted to win.
// The results are then sorted by win probability and returned in the JSON response.
// Helper Functions:

// parseDistance: Placeholder for converting distance strings to numeric values. You should implement this based on your data format.
// parseOdds: Converts fractional odds to decimal odds. Modify this function according to the odds format you are using.

package analysis

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/mmanjoura/race-picks-backend/pkg/database"
// )

// type Selection struct {
// 	ID        int64  `json:"id"`
// 	Name      string `json:"name"`
// 	EventName string `json:"event_name"`
// 	EventTime string `json:"event_time"`
// 	Distance  string `json:"distance"`
// 	Odds      string `json:"odds"`
// }

// type SVMPrediction struct {
// 	SelectionID    int64   `json:"selection_id"`
// 	SelectionName  string  `json:"selection_name"`
// 	EventName      string  `json:"event_name"`
// 	EventTime      string  `json:"event_time"`
// 	PredictedWin   bool    `json:"predicted_win"`
// 	WinProbability float64 `json:"win_probability"`
// }

// func SVM_Prediction(c *gin.Context) {
// 	db := database.Database.DB
// 	var modelparams Selection

// 	if err := c.ShouldBindJSON(&modelparams); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Get today's runners for the given event_name and event_date
// 	rows, err := db.Query(`
// 		SELECT selection_id, 
// 			   selection_name,
// 			   event_name,
// 			   event_time,
// 			   distance,
// 			   odds
// 		FROM TodayRunners
// 		WHERE event_name = ? AND DATE(event_date) = DATE('now') AND event_time = ?`,
// 		modelparams.EventName, modelparams.EventTime)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	var selections []Selection
// 	for rows.Next() {
// 		var selection Selection
// 		if err := rows.Scan(&selection.ID, &selection.Name, &selection.EventName, &selection.EventTime, &selection.Distance, &selection.Odds); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		selections = append(selections, selection)
// 	}

// 	// Prepare data for external SVM service
// 	svmData := make([]map[string]interface{}, len(selections))
// 	for i, selection := range selections {
// 		svmData[i] = map[string]interface{}{
// 			"selection_id": selection.ID,
// 			"distance":     parseDistance(selection.Distance),
// 			"odds":         parseOdds(selection.Odds),
// 		}
// 	}

// 	// Send data to SVM prediction service
// 	svmURL := "http://localhost:5000/predict_svm"
// 	data, err := json.Marshal(svmData)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	resp, err := http.Post(svmURL, "application/json", bytes.NewBuffer(data))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer resp.Body.Close()

// 	var predictions []struct {
// 		Probability float64 `json:"probability"`
// 	}
// 	if err := json.NewDecoder(resp.Body).Decode(&predictions); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Process predictions and prepare results
// 	var results []SVMPrediction
// 	for i, selection := range selections {
// 		results = append(results, SVMPrediction{
// 			SelectionID:    selection.ID,
// 			SelectionName:  selection.Name,
// 			EventName:      selection.EventName,
// 			EventTime:      selection.EventTime,
// 			PredictedWin:   predictions[i].Probability > 0.5, // Example threshold for classification
// 			WinProbability: predictions[i].Probability,
// 		})
// 	}

// 	// Sort results by win probability
// 	sort.SliceStable(results, func(i, j int) bool {
// 		return results[i].WinProbability > results[j].WinProbability
// 	})

// 	c.JSON(http.StatusOK, gin.H{"data": results})
// }

// // Helper function to parse distance string to a numeric value (e.g., yards)
// func parseDistance(distanceStr string) float64 {
// 	// Implement your logic to convert distance from string to a numeric value (e.g., in yards)
// 	// For example, return a mock value for now
// 	return 0.0
// }

// // Helper function to parse odds string to a numeric value (e.g., decimal odds)
// func parseOdds(oddsStr string) float64 {
// 	// Implement your logic to convert odds from string to a numeric value
// 	// For example, convert odds from fractional to decimal format
// 	if strings.Contains(oddsStr, "/") {
// 		parts := strings.Split(oddsStr, "/")
// 		if len(parts) == 2 {
// 			numerator, err1 := strconv.ParseFloat(parts[0], 64)
// 			denominator, err2 := strconv.ParseFloat(parts[1], 64)
// 			if err1 == nil && err2 == nil && denominator != 0 {
// 				return numerator / denominator + 1
// 			}
// 		}
// 	}
// 	return 0.0
// }
