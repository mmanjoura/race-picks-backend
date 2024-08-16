package analysis

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mmanjoura/race-picks-backend/pkg/api/common"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
)

type SimulationResult struct {
	SelectionID    int64   `json:"selection_id"`
	SelectionName  string  `json:"selection_name"`
	EventName      string  `json:"event_name"`
	EventDate      string  `json:"event_date"`
	EventTime      string  `json:"event_time"`
	Odds           string  `json:"odds"`
	WinProbability float64 `json:"win_probability"`
}

func MonteCarloSimulation(c *gin.Context) {
	db := database.Database.DB
	var modelparams common.Selection

	if err := c.ShouldBindJSON(&modelparams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get today's runners for the given event_name and event_date
	rows, err := db.Query(`
		SELECT selection_id, 
			   selection_name,
			   event_date,
			   event_name,
			   event_time,
			   price
	
		FROM EventRunners	 
		WHERE event_name = ? AND DATE(event_date) = ? AND event_time = ?`,
		modelparams.EventName, modelparams.EventDate, modelparams.EventTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var selections []common.Selection
	for rows.Next() {
		var selection common.Selection
		if err := rows.Scan(&selection.ID, &selection.Name, selection.EventDate, &selection.EventName, &selection.EventTime, &selection.Odds); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		selections = append(selections, selection)
	}

	// Create a new random generator with a seed
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Number of simulations
	const numSimulations = 10000

	// Simulate the race multiple times
	results := make(map[int64]int)
	for i := 0; i < numSimulations; i++ {
		winner := simulateRace(selections, db, rng, modelparams.Distance)
		results[winner]++
	}

	// Calculate win probabilities
	var simulationResults []SimulationResult
	for _, selection := range selections {
		winCount := results[selection.ID]
		winProbability := float64(winCount) / float64(numSimulations)
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

	sort.SliceStable(simulationResults, func(i, j int) bool {
		return simulationResults[i].WinProbability > simulationResults[j].WinProbability
	})

	// Sort results by win probability (optional)

	c.JSON(http.StatusOK, gin.H{"data": simulationResults})
}

func simulateRace(selections []common.Selection, db *sql.DB, rng *rand.Rand, eventDistance string) int64 {
	// Define probabilities for each selection based on historical data
	probabilities := make(map[int64]float64)

	for _, selection := range selections {
		probability := calculateProbability(selection.ID, eventDistance, db)
		probabilities[selection.ID] = probability
	}

	// Randomly pick a winner based on the calculated probabilities
	totalProb := 0.0
	for _, prob := range probabilities {
		totalProb += prob
	}
	threshold := rng.Float64() * totalProb
	cumulative := 0.0
	for id, prob := range probabilities {
		cumulative += prob
		if threshold <= cumulative {
			return id
		}
	}

	// Fallback: return the last one if something went wrong
	return selections[len(selections)-1].ID
}

func calculateProbability(selectionID int64, distance string, db *sql.DB) float64 {
	// Query historical data for this selection, including all parameters
	rows, err := db.Query(`
		SELECT 
			COUNT(*) AS count, 
			position, 
			rating, 
			distance, 
			sp_odds,
			AVG(recovery_days) AS avg_recovery_days,
			AVG(num_runs) AS avg_num_runs,
			AVG(years_running) AS avg_years_running,
			AVG(win_count) AS avg_win_count,
			AVG(avg_position) AS avg_avg_position,
			AVG(avg_distance_furlongs) AS avg_avg_distance_furlongs
		FROM SelectionsForm
		WHERE selection_id = ?
		GROUP BY selection_id, position, rating, distance, sp_odds
	`, selectionID)
	if err != nil {
		fmt.Println("Error querying historical data:", err)
		return 0.0
	}
	defer rows.Close()

	var totalRuns int
	var totalScore float64

	for rows.Next() {
		var positionStr, distanceStr string
		var rating, count int
		var odds float64
		var avgRecoveryDays, avgNumRuns, avgYearsRunning, avgWinCount, avgAvgPosition, avgAvgDistanceFurlongs float64

		if err := rows.Scan(
			&count,
			&positionStr,
			&rating,
			&distanceStr,
			&odds,
			&avgRecoveryDays,
			&avgNumRuns,
			&avgYearsRunning,
			&avgWinCount,
			&avgAvgPosition,
			&avgAvgDistanceFurlongs); err != nil {
			fmt.Println("Error scanning historical data:", err)
			continue
		}

		// Parse the position string (e.g., "3/11" -> 3)
		positionParts := strings.Split(positionStr, "/")
		if len(positionParts) > 0 {
			// Check if the position is a numeric value
			position, err := strconv.Atoi(positionParts[0])
			if err != nil {
				fmt.Println("Non-numeric position encountered, skipping:", positionParts[0])
				continue
			}
			totalRuns++
			// Invert position to score (lower position = higher score)
			totalScore += 1 / float64(position)
		}

		// Factor in other parameters to adjust the score
		// You can adjust the weight of each parameter according to its significance
		totalScore *= (avgRecoveryDays + avgNumRuns + avgYearsRunning + avgWinCount + avgAvgPosition + avgAvgDistanceFurlongs)

		// Compare distances (use a scaling factor based on distance difference)
		historicalDistance := common.ParseDistance(distanceStr)
		currentDistance := common.ParseDistance(distance)

		// Weight by how close the historical distance is to the current distance
		distanceWeight := 1.0 / (1.0 + float64(common.Abs(historicalDistance-currentDistance))/1000.0)
		totalScore *= distanceWeight
	}

	if totalRuns == 0 {
		return 0.0
	}

	// Normalize the score to be used as a probability
	return totalScore / float64(totalRuns)
}
