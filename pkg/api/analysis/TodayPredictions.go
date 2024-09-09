package analysis

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"

	"net/http"

	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

// RacePicksSimulation handles the simulation of race picks and calculates win probabilities.
func GetTodayPredictions(c *gin.Context) {
	db := database.Database.DB
	var raceParams models.RaceParameters

	// Bind JSON input to optimalParams
	if err := c.ShouldBindJSON(&raceParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get all meeting of today
	var eventNames []string
	rows, err := db.Query(`SELECT DISTINCT event_name FROM EventRunners WHERE DATE(event_date) = ? `, raceParams.EventDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var eventName string
		err := rows.Scan(&eventName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		eventNames = append(eventNames, eventName)
	}

	// Get Event Times for the given event names

	// Create map of event names to event times
	var eventTimesMap = make(map[string][]string)
	for _, eventName := range eventNames {
		var eventTimes []string
		rows, err = db.Query(`SELECT DISTINCT event_time FROM EventRunners WHERE event_name = ? AND DATE(event_date) = ?`, eventName, raceParams.EventDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		for rows.Next() {
			var eventTime string
			err := rows.Scan(&eventTime)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			eventTimes = append(eventTimes, eventTime)
		}
		eventTimesMap[eventName] = eventTimes
	}

	// loop through eventTimesMap
	for eventName, eventTimes := range eventTimesMap {
		for _, eventTime := range eventTimes {
			raceParams.EventName = eventName
			raceParams.EventTime = eventTime

			// Convert the struct to JSON
			jsonData, err := json.Marshal(raceParams)
			if err != nil {
				fmt.Println("Error encoding JSON:", err)
				return
			}

			// Create a new POST request with the JSON body
			req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/analysis/MeetingPrediction", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println("Error creating request:", err)
				return
			}

			// Set the appropriate headers
			req.Header.Set("Content-Type", "application/json")

			// Send the request using the default HTTP client
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Error sending request:", err)
				return
			}
			defer resp.Body.Close()

		}
	}

	c.JSON(http.StatusOK, gin.H{"simulationResults": "Predictions have been made successfully"})
}


func insertPredictions(db *sql.DB, data models.SelectionResult) error {

	// Prepare the INSERT statement
	stmt, err := db.Prepare(`
		INSERT INTO EventPredictions (event_date, selection_id, selection_name, odds, clean_bet_score, average_position, average_rating, event_name, event_time, num_runners)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the INSERT statement
	_, err = stmt.Exec(data.EventDate, data.SelectionID, data.SelectionName, data.Odds, data.TotalScore, data.AvgPosition, data.AvgRating, data.EventName, data.EventTime, data.RunCount)
	if err != nil {
		return err
	}

	// Return nil if no error occurred
	return nil
}
func deletePredictions(db *sql.DB, eventDate, eventName, eventTime string) error {
	// Check if a record with the same event_date and selection_id exists
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM EventPredictions
			WHERE DATE(event_date) = ? and event_name = ? and event_time = ?
		)
	`, eventDate, eventName, eventTime).Scan(&exists)
	if err != nil {
		return err
	}

	// If a record exists, delete it
	if exists {
		_, err = db.Exec(`
			DELETE FROM EventPredictions
			WHERE DATE(event_date) = ? and event_name = ? and event_time = ?
		`, eventDate, eventName, eventTime)
		if err != nil {
			return err
		}
	}

	// Return nil if no error occurred
	return nil
}

// FindBestSelection returns the selection with the highest score, highest rating, and youngest age
func FindBestSelection(data []models.SelectionResult) (models.SelectionResult, error) {
	var bestSelection models.SelectionResult
	var found bool

	// Iterate through the map to find the best selection

	for _, record := range data {
		// Extract age from the name (e.g., "Little Empire 4" -> 4)
		age := extractAge(record.SelectionName)

		// Compare based on the criteria: highest score, highest rating, and youngest age
		if !found || isBetterSelection(record, bestSelection, age) {
			bestSelection = record
			bestSelection.Age = strconv.Itoa(age) // Convert age to string and assign it
			found = true
		}
	}

	if !found {
		return models.SelectionResult{}, fmt.Errorf("no selections found")
	}

	return bestSelection, nil
}

// Extracts the age from the name of the selection
func extractAge(name string) int {
	// Split the name by space and get the last part, assuming it's the age
	parts := strings.Split(name, " ")
	age := 0
	if len(parts) > 0 {
		fmt.Sscanf(parts[len(parts)-1], "%d", &age) // Read age as integer
	}
	return age
}

// Checks if the current selection is better based on the criteria
func isBetterSelection(current, best models.SelectionResult, currentAge int) bool {
	bestAge, _ := strconv.Atoi(best.Age) // Convert best.Age from string to int
	if current.TotalScore > best.TotalScore {
		return true
	} else if current.TotalScore == best.TotalScore {
		if current.AvgRating > best.AvgRating {
			return true
		} else if current.AvgRating == best.AvgRating {
			if currentAge < bestAge { // Compare currentAge with bestAge
				return true
			}
		}
	}
	return false
}
