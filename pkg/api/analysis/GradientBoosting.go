// Explanation
// Data Preparation:

// The handler first gathers the necessary data from the database.
// It then formats this data into a structure suitable for sending to the GBM service.
// External GBM Service:

// The handler sends a POST request with the formatted data to an external GBM service running on localhost:5000.
// The Python REST API is expected to return predictions as a list of probabilities.
// Processing Predictions:

// After receiving the predictions, the handler maps these probabilities back to the original selections.
// It uses a threshold to determine if a selection is predicted to win (e.g., a probability greater than 0.5).
// The results are sorted by win probability and returned as a JSON response.
// Helper Functions:

// parseDistance: Placeholder for converting distance strings to numeric values. You should implement the actual conversion logic.
// parseOdds: Converts fractional odds to decimal odds. Adjust according to the odds format you are using.
// Notes
// External Service: Ensure your external GBM service is running and accessible from your Go application. Modify the URL (http://localhost:5000/predict) as needed.
// Error Handling: Proper error handling and logging are crucial for production use. The provided code includes basic error handling.
// Distance and Odds Parsing: Adjust these helper functions based on the actual format and requirements of your data.

package analysis

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http"
// 	"strconv"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// )

// type GBMSelection struct {
// 	ID        int64  `json:"id"`
// 	Name      string `json:"name"`
// 	EventName string `json:"event_name"`
// 	EventTime string `json:"event_time"`
// 	Distance  string `json:"event_distance"`
// 	Odds      string `json:"odds"`
// }

// type GBMPrediction struct {
// 	SelectionID    int64   `json:"selection_id"`
// 	SelectionName  string  `json:"selection_name"`
// 	EventName      string  `json:"event_name"`
// 	EventTime      string  `json:"event_time"`
// 	PredictedWin   bool    `json:"predicted_win"`
// 	WinProbability float64 `json:"win_probability"`
// }

// // GradientBoostingPrediction predicts the winner using Gradient Boosting Machines
// func GradientBoostingPrediction(c *gin.Context) {
// 	var modelparams GBMSelection

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
// 			   price
// 		FROM TodayRunners	 
// 		WHERE event_name = ? AND DATE(event_date) = DATE('now') AND event_time = ?`,
// 		modelparams.EventName, modelparams.EventTime)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	var selections []GBMSelection
// 	for rows.Next() {
// 		var selection GBMSelection
// 		if err := rows.Scan(&selection.ID, &selection.Name, &selection.EventName, &selection.EventTime, &selection.Odds); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		selections = append(selections, selection)
// 	}

// 	// Prepare data for GBM prediction
// 	data := make

// 	defer resp.Body.Close()

// 	var predictions []float64
// 	if err := json.NewDecoder(resp.Body).Decode(&predictions); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Process predictions and prepare results
// 	var results []GBMPrediction
// 	for i, selection := range selections {
// 		results = append(results, GBMPrediction{
// 			SelectionID:    selection.ID,
// 			SelectionName:  selection.Name,
// 			EventName:      selection.EventName,
// 			EventTime:      selection.EventTime,
// 			PredictedWin:   predictions[i] > 0.5, // Example threshold for classification
// 			WinProbability: predictions[i],
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

