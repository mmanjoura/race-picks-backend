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
	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/api/common"
)


type SimulationResult struct {
	SelectionID    int64   `json:"selection_id"`
	SelectionName  string  `json:"selection_name"`
	EventName      string  `json:"event_name"`
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
			   event_name,
			   event_time,
			   price
	
		FROM EventRunners	 
		WHERE event_name = ? AND DATE(event_date) = DATE('now') AND event_time = ?`,
		modelparams.EventName, modelparams.EventTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var selections []common.Selection
	for rows.Next() {
		var selection common.Selection
		if err := rows.Scan(&selection.ID, &selection.Name, &selection.EventName, &selection.EventTime, &selection.Odds); err != nil {
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
	// Query historical data for this selection
	var totalRuns int
	var totalScore float64

	rows, err := db.Query(`
		SELECT COUNT(*) AS count, position, rating, distance 
			FROM SelectionsForm
			WHERE selection_id = ?
			GROUP BY selection_id, position, rating, distance
			`, selectionID)
	if err != nil {
		fmt.Println("Error querying historical data:", err)
		return 0.0
	}
	defer rows.Close()

	for rows.Next() {
		var positionStr, distanceStr string
		var rating, count int
		if err := rows.Scan(&count, &positionStr, &rating, &distanceStr); err != nil {
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
