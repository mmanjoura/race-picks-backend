package preparation

import (
	"database/sql"
	"net/http"
	"sort"
	"strings"

	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

func GetPredictions(c *gin.Context) {
	db := database.Database.DB

	// Query for today's runners
	date := c.Query("event_date")
	currentBet := 10.0
	totalBet := 10.0
	var eventPredicitonsResponse models.EventPredictionResponse

	var pnl []float64
	var totalBets []float64

	rows, err := db.Query("SELECT starting_amount, current_amount, profit_loss FROM User")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var startingAmount, currentAmount, profitLoss sql.NullFloat64
	for rows.Next() {
		err := rows.Scan(&startingAmount, &currentAmount, &profitLoss)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	rows, err = db.Query(`SELECT id,
							selection_id,
							selection_name,
							odds,
							age,
							clean_bet_score,
							average_position,
							average_rating,
							event_name,
							event_date,
							race_date,
							event_time,
							selection_position,
							ABS(prefered_distance - current_distance)  as distanceTolerence,
							num_runners,
							number_runs,
							prefered_distance,
							current_distance,
							created_at,
							updated_at
						FROM EventPredictions
						WHERE event_date = ? and  distanceTolerence < 1 and average_position < 2 and number_runs < 10 
						order by clean_bet_score DESC`, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var predictions []models.EventPrediction
	for rows.Next() {
		racePrdiction := models.EventPrediction{}
		err := rows.Scan(
			&racePrdiction.ID,
			&racePrdiction.SelectionID,
			&racePrdiction.SelectionName,
			&racePrdiction.Odds,
			&racePrdiction.Age,
			&racePrdiction.CleanBetScore,
			&racePrdiction.AveragePosition,
			&racePrdiction.AverageRating,
			&racePrdiction.EventName,
			&racePrdiction.EventDate,
			&racePrdiction.RaceDate,
			&racePrdiction.EventTime,
			&racePrdiction.SelectionPosition,
			&racePrdiction.DistanceTolerence,
			&racePrdiction.NumRunners,
			&racePrdiction.NumbeRuns,
			&racePrdiction.PreferredDistance,
			&racePrdiction.CurrentDistance,
			&racePrdiction.CreatedAt,
			&racePrdiction.UpdatedAt,
			
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		position, err := getPosition(racePrdiction.SelectionID, date, db)

		if err != nil {
			continue
		}

		racePrdiction.Position = position


		predictions = append(predictions, racePrdiction)

	}
	
	// First filter the predictions to remove duplicates based on EventTime
	filteredPredictions := filterHighestBetScore(predictions)

	// After filtering, compute PnL
	for _, p := range filteredPredictions {		
		
		position := strings.Split(p.Position, "/")[0]
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if position == "1"{
			pnl = append(pnl, currentBet*p.Odds)
		}
		totalBets = append(totalBets, totalBet)
	}

	TotalBet := sumSlice(totalBets)
	TotalReturn := sumSlice(pnl)

	eventPredicitonsResponse.Selections = filteredPredictions
	eventPredicitonsResponse.TotalBet = TotalBet
	eventPredicitonsResponse.TotalReturn = TotalReturn - TotalBet

	// Sort filtered predictions by CleanBetScore if needed (descending order)
	sort.Slice(filteredPredictions, func(i, j int) bool {
		return filteredPredictions[i].CleanBetScore > filteredPredictions[j].CleanBetScore
	})

	c.JSON(http.StatusOK, gin.H{"predictions": eventPredicitonsResponse})
}

func sumSlice(slice []float64) float64 {
	var sum float64
	for _, value := range slice {
		sum += value
	}
	return sum
}

func filterHighestBetScore(predictions []models.EventPrediction) []models.EventPrediction {
	// Create a map to store the highest CleanBetScore for each EventTime
	eventTimeMap := make(map[string]models.EventPrediction)

	// Iterate through predictions and keep only the one with the highest CleanBetScore for each EventTime
	for _, prediction := range predictions {
		existing, found := eventTimeMap[prediction.EventTime]
		if !found || prediction.CleanBetScore > existing.CleanBetScore {
			eventTimeMap[prediction.EventTime] = prediction
		}
	}

	// Convert map to a slice of EventPredictions
	filteredPredictions := make([]models.EventPrediction, 0, len(eventTimeMap))
	for _, prediction := range eventTimeMap {
		filteredPredictions = append(filteredPredictions, prediction)
	}

	return filteredPredictions
}

// Get Postion given selection Id
func getPosition(selectionId int, race_date string, db *sql.DB) (string, error) {
	var position string
	err := db.QueryRow(`
				Select position 
				from selectionsForm 
				where DATE(race_date) = ? and selection_id = ?;`, race_date, selectionId).Scan(&position)
	if err != nil {
		return "", err
	}
	return position, nil
}