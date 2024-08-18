package analysis

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mmanjoura/race-picks-backend/pkg/api/common"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

// RacePicksSimulation handles the simulation of race picks and calculates win probabilities.
func RacePicksSimulation(c *gin.Context) {
	db := database.Database.DB
	var optimalParams models.OptimalParameters

	// Bind JSON input to optimalParams
	if err := c.ShouldBindJSON(&optimalParams); err != nil {
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
		WHERE event_name = ? AND DATE(event_date) = ? AND event_time = ?`,
		optimalParams.EventName, optimalParams.EventDate, optimalParams.EventTime)
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
	var simulationResults []SimulationResult
	for _, selection := range selections {
		selectionRows, err := db.Query(`
			SELECT
				selection_name,
				COALESCE(COUNT(*), 0) AS NumberOfRuns,
				COALESCE(MAX(race_date) - MIN(race_date), 0) AS YearsRunning,
				COALESCE(COUNT(CASE WHEN position = '1' THEN 1 END), 0) AS NumberOfWins,
				COALESCE(AVG(position), 0) AS AveragePosition,
				COALESCE(AVG(rating), 0) AS AverageRating,
				COALESCE(AVG(distance), 0) AS AverageDistance,
				COALESCE(AVG(sp_odds), 0) AS AverageOdds
			FROM
				SelectionsForm
			WHERE
				selection_id = ?
			ORDER BY
				race_date DESC`, selection.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer selectionRows.Close()

		var horseParams models.CurrentHorseData
		for selectionRows.Next() {
			if err := selectionRows.Scan(
				&horseParams.Name,
				&horseParams.NumberOfRuns,
				&horseParams.YearsRunning,
				&horseParams.NumberOfWins,
				&horseParams.AveragePosition,
				&horseParams.AverageRating,
				&horseParams.AverageDistance,
				&horseParams.AverageOdds); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Calculate the win probability
			winProbability := CalculateWinProbability(horseParams, optimalParams)
			if winProbability == 1.0 {
				simulationResults = append(simulationResults, SimulationResult{
					SelectionID:    selection.ID,
					SelectionName:  selection.Name,
					EventName:      selection.EventName,
					EventDate:      selection.EventDate,
					EventTime:      selection.EventTime,
					Odds:           selection.Odds,
					WinProbability: winProbability,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"simulationResults": simulationResults})
}

// CalculateWinProbability calculates the probability of winning based on historical data and optimal parameters.
func CalculateWinProbability(horseParams models.CurrentHorseData, optimalParams models.OptimalParameters) float64 {

	// If all true
	if IsWithinTolerance(float64(horseParams.NumberOfRuns), float64(optimalParams.OptimalNumRuns), optimalParams.Tolerance) &&
		IsWithinTolerance(float64(horseParams.YearsRunning), float64(optimalParams.OptimalNumYearsInCompetition), optimalParams.Tolerance) &&
		IsWithinTolerance(float64(horseParams.NumberOfWins), float64(optimalParams.OptimalNumWins), optimalParams.Tolerance) &&
		IsWithinTolerance(horseParams.AverageRating, optimalParams.OptimalRating, optimalParams.Tolerance * 10.0) &&
		IsWithinTolerance(horseParams.AveragePosition, optimalParams.OptimalPosition, optimalParams.Tolerance) &&
		IsWithinTolerance(horseParams.AverageDistance, optimalParams.OptimalDistance, optimalParams.Tolerance) {
		return 1.0
	}

	return 0.0
}

func IsWithinTolerance(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
