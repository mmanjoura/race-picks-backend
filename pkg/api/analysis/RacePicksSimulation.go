package analysis

import (
	"math"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/mmanjoura/race-picks-backend/pkg/api/common"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
)

// Horse represents the data structure for a horse's history.
type Horse struct {
	DaysSinceLastRun int
	NumberOfRuns     int
	YearsRunning     int
	NumberOfWins     int
	AverageRating    float64
	AveragePosition  float64
	AverageDistance  float64
}

func RacePicksSimulation(c *gin.Context) {
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
		if err := rows.Scan(&selection.ID, &selection.Name, &selection.EventName, &selection.EventTime, &selection.Odds); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		selections = append(selections, selection)
	}



	// Calculate win probabilities
	var simulationResults []SimulationResult
	for _, selection := range selections {
	
		simulationResults = append(simulationResults, SimulationResult{
			SelectionID:    selection.ID,
			SelectionName:  selection.Name,
			EventName:      selection.EventName,
			EventTime:      selection.EventTime,
			Odds:           selection.Odds,
			// WinProbability: winProbability,
		})
	}

	sort.SliceStable(simulationResults, func(i, j int) bool {
		return simulationResults[i].WinProbability > simulationResults[j].WinProbability
	})

	// Sort results by win probability (optional)

	c.JSON(http.StatusOK, gin.H{"data": simulationResults})
}




// CalculateWinProbability calculates the probability
// of the horse winning based on given parameters.
func CalculateWinProbability(horse Horse) float64 {

	// Normalize each factor to a scale of 0 to 1 (1 being the best possible scenario for winning).
	// Adjust weights according to their importance in predicting the win.

	// Normalize DaysSinceLastRun (assuming optimal range is 20-30 days)
	daysSinceLastRunFactor := math.Max(0, math.Min(1, 1-math.Abs(float64(horse.DaysSinceLastRun-14))/7))

	// Normalize NumberOfRuns (more experience is usually better)
	// Assume a plateau effect where >50 runs doesn't significantly increase the chance further
	numberOfRunsFactor := math.Min(1, float64(horse.NumberOfRuns)/30)

	// Normalize YearsRunning (more years could indicate more experience or more wear)
	yearsRunningFactor := math.Min(1, float64(horse.YearsRunning)/5)

	// Normalize NumberOfWins (the more wins, the better but no more than 6 wins)
	// We map the win rate to a scale between 0 and 6 and cap the value at 6.
	numberOfWinsFactor := 0.0

	if horse.NumberOfRuns > 0 {
		// Calculate win rate (between 0 and 1)
		winRate := float64(horse.NumberOfWins) / float64(horse.NumberOfRuns)

		// Scale win rate to a value between 0 and 6
		numberOfWinsFactor = winRate * 6

		// Cap the factor at 6
		if numberOfWinsFactor > 6 {
			numberOfWinsFactor = 6
		}
	}
	// Normalize AverageRating (higher rating is better, assuming ratings go from 0 to 100)
	averageRatingFactor := horse.AverageRating / 50

	// Normalize AveragePosition (lower positions are better; position 1 is the best)
	averagePositionFactor := 1 / horse.AveragePosition

	// Normalize AverageDistance (assume an optimal distance range based on the horse's typical run)
	optimalDistance := 1600.0 // Example: optimal distance might be 1600 meters (1 mile)
	averageDistanceFactor := math.Max(0, math.Min(1, 1-math.Abs(horse.AverageDistance-optimalDistance)/800))

	// Assign weights to each factor (these can be adjusted)
	weights := []float64{0.15, 0.10, 0.10, 0.20, 0.20, 0.15, 0.10}
	factors := []float64{
		daysSinceLastRunFactor,
		numberOfRunsFactor,
		yearsRunningFactor,
		numberOfWinsFactor,
		averageRatingFactor,
		averagePositionFactor,
		averageDistanceFactor,
	}

	// Calculate the weighted sum of factors
	probability := 0.0
	for i, factor := range factors {
		probability += factor * weights[i]
	}

	// Return probability as a percentage
	return probability * 100
}

// func main() {
// 	// Example horse data
// 	horse := Horse{
// 		DaysSinceLastRun: 14,
// 		NumberOfRuns:     25,
// 		YearsRunning:     3,
// 		NumberOfWins:     8,
// 		AverageRating:    85.0,
// 		AveragePosition:  3.2,
// 		AverageDistance:  1600.0,
// 	}

// 	// Calculate and print the probability of winning
// 	probability := CalculateWinProbability(horse)
// 	fmt.Printf("The probability of the horse winning is %.2f%%\n", probability)
// }
