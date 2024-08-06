// Explanation
// Data Preparation:

// Fetch selection data from the database based on the event name and time.
// Format the data for sending to the external ML service. This might include normalization or 
// transformation to match what the ML model expects.
// External ML Model Service:

// Send the formatted data to an external ML model service using a POST request. 
// The URL (http://localhost:5000/predict_ml) should be replaced with the actual endpoint where your model is served.
// Processing Predictions:

// Parse the response from the ML service to extract predictions.
// Map these predictions to the selections and determine if each selection is predicted to win based 
// on a threshold (e.g., 0.5 for binary classification).
// Sort the results by win probability and respond with the results.
// Helper Functions:

// parseDistance: Convert distance strings to numeric values based on your specific format.
// parseOdds: Convert fractional odds to decimal odds or other formats as needed.
// Notes
// External Service: Ensure the external ML model service is set up and accessible from your Go application.
// Data Conversion: Adjust the data parsing functions based on the actual format of the input data and predictions.
// Error Handling: Basic error handling is implemented. Add more detailed handling as needed for production environments.
// This handler integrates with an external machine learning model to perform predictions, 
// allowing you to leverage advanced ML techniques without having to implement the model directly in Go.

package analysis

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"strconv"
// 	"strings"

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

// type MLModelPrediction struct {
// 	SelectionID    int64   `json:"selection_id"`
// 	SelectionName  string  `json:"selection_name"`
// 	EventName      string  `json:"event_name"`
// 	EventTime      string  `json:"event_time"`
// 	PredictedWin   bool    `json:"predicted_win"`
// 	WinProbability float64 `json:"win_probability"`
// }

// func MLModel_Prediction(c *gin.Context) {
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

// 	// Prepare data for external ML model service
// 	mlData := make([]map[string]interface{}, len(selections))
// 	for i, selection := range selections {
// 		mlData[i] = map[string]interface{}{
// 			"selection_id": selection.ID,
// 			"distance":     parseDistance(selection.Distance),
// 			"odds":         parseOdds(selection.Odds),
// 		}
// 	}

// 	// Send data to ML model prediction service
// 	mlURL := "http://localhost:5000/predict_ml"
// 	data, err := json.Marshal(mlData)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	resp, err := http.Post(mlURL, "application/json", bytes.NewBuffer(data))
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
// 	var results []MLModelPrediction
// 	for i, selection := range selections {
// 		results = append(results, MLModelPrediction{
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
