package analysis

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mmanjoura/race-picks-backend/pkg/api/common"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

// RacePicksSimulation handles the simulation of race picks and calculates win probabilities.
func RacePicksSimulation(c *gin.Context) {
	db := database.Database.DB
	var raceParams models.RaceParameters

	// Bind JSON input to optimalParams
	if err := c.ShouldBindJSON(&raceParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Query for today's runners
	rows, err := db.Query(`
		SELECT selection_id, 
			   selection_name,
			   event_name,
			   event_date,
			   event_time,
			   price
		FROM EventRunners	 
		WHERE DATE(event_date) = ? and  event_time = ?`,
		raceParams.EventDate, raceParams.EventTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var selections []common.Selection
	for rows.Next() {
		var selection common.Selection
		if err := rows.Scan(&selection.ID, &selection.Name, &selection.EventName, &selection.EventDate, &selection.EventTime, &selection.Odds); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		selections = append(selections, selection)
	}

	// Calculate win probabilities
	analysisData := make(map[string]models.AnalysisData) // Map to store the highest rated selection for each event

	for _, selection := range selections {
		// Execute the query
		rows, err = db.Query(`
				SELECT
					COALESCE(selection_id, 0),
					selection_name,	
					substr(position, 1, 1) as positon, 
					Age,
					Trainer,
					Sex,
					Sire,
					Dam,
					Owner,						
					COUNT(*) AS num_runs,
					MAX(race_date) AS last_run_date,
					MAX(race_date) - MIN(race_date) AS duration,
					COUNT(CASE WHEN position = '1' THEN 1 END) AS win_count,
					AVG(position) AS avg_position,
					AVG(rating) AS avg_rating,
					AVG(distance) AS avg_distance_furlongs,
					AVG(sp_odds) AS sp_odds,
					GROUP_CONCAT(position, ', ') AS all_positions,
					GROUP_CONCAT(distance, ', ') AS all_distances,
					GROUP_CONCAT(racecourse, ', ') AS all_racecources,
					GROUP_CONCAT(DATE(race_date), ', ') AS all_race_dates 
				FROM
					SelectionsForm	WHERE selection_id = ?  order by race_date desc`, selection.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var data models.AnalysisData
		for rows.Next() {
			err := rows.Scan(
				&data.SelectionID,
				&data.SelectionName,
				&data.Position,
				&data.Age,
				&data.Trainer,
				&data.Sex,
				&data.Sire,
				&data.Dam,
				&data.Owner,
				&data.NumRuns,
				&data.LastRunDate,
				&data.Duration,
				&data.WinCount,
				&data.AvgPosition,
				&data.AvgRating,
				&data.AvgDistanceFurlongs,
				&data.AvgOdds,
				&data.AllPositions,
				&data.AllDistances,
				&data.AllCources,
				&data.AllRaceDates,
			)
			if err != nil {
				continue
			}
			data.EventName = selection.EventName
			data.EventTime = selection.EventTime
			data.EventDate = selection.EventDate

			if getPotentialWinners(data) {
				// Check if the event already has a winner with higher average rating
				if existingData, exists := analysisData[data.EventName]; !exists || data.AvgRating > existingData.AvgRating {
					analysisData[data.EventName] = data // Replace or add the new winner with the higher rating
				}
			}
		}
	}

	// Convert the map to a slice for the final result
	finalResults := make([]models.AnalysisData, 0, len(analysisData))
	for _, winner := range analysisData {
		finalResults = append(finalResults, winner)
	}

	c.JSON(http.StatusOK, gin.H{"simulationResults": finalResults})
}

func IsWithinTolerance(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func getPotentialWinners(analysisData models.AnalysisData) bool {
	// Step 3: Check NumRuns < 8
	if analysisData.NumRuns < 10 && analysisData.AvgPosition < 4 {
		// Step 1: Filter AllRaceDates to exclude years 2022, 2021, 2020, 2019, 2018, 2017
		excludedYears := []int{2022, 2021, 2020, 2019, 2018, 2017}
		doesIncludeYears := includedRaceDates(analysisData.AllRaceDates, excludedYears)

		// Step 2: Filter AllPositions to exclude positions starting with 1/*, 7/*, 8/*, 9/*, 10/*, 11/*, 12/*, 13/*, 14/*, 15/*, 16/*, 17/*
		excludedPatterns := []string{"7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17"}
		doesIncludePositions := includedPositions(analysisData.AllPositions, excludedPatterns)

		if !doesIncludeYears && !doesIncludePositions {
			fmt.Printf("Filtered AnalysisData: %+v\n", analysisData)
			return true
		}
	}
	return false
}

// includedRaceDates checks if any race dates fall within the excluded years.
func includedRaceDates(dates string, excludedYears []int) bool {
	var dateIncluded bool
	for _, date := range strings.Split(dates, ", ") {
		if d, err := time.Parse("2006-01-02", date); err == nil {
			year := d.Year()
			if containsInt(excludedYears, year) {
				dateIncluded = true
			}
		}
	}
	return dateIncluded
}

// includedPositions checks if positions include any excluded patterns.
func includedPositions(positions string, excludedPatterns []string) bool {
	positionIncluded := false
	for _, position := range strings.Split(positions, ", ") {
		p := strings.Split(position, "/")[0]
		if containString(excludedPatterns, p) {
			positionIncluded = true
		}
	}
	return positionIncluded
}

// containString checks if a string slice contains a given string.
func containString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// containsInt checks if an int slice contains a given int.
func containsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
