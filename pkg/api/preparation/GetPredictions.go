package preparation

import (
	"database/sql"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

func GetPredictions(c *gin.Context) {
	db := database.Database.DB

	params := models.GetWinnerParams{}

	// Bind JSON input to optimalParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
							COALESCE(odds, '') as odds,
							COALESCE(age, '') as age,
							COALESCE(clean_bet_score, '') as clean_bet_score,
							COALESCE(average_position, '') as average_position,
							COALESCE(average_rating, '') as average_rating,
							event_name,
							COALESCE(event_date, '') as event_date,
							COALESCE(race_date, '') as race_date,
							COALESCE(event_time, '') as event_time,
							COALESCE(selection_position, '') as selection_position,
							ABS(prefered_distance - current_distance) as distanceTolerence,
							COALESCE(num_runners, '') as num_runners,
							COALESCE(number_runs, '') as number_runs,
							COALESCE(prefered_distance, '') as prefered_distance,
							COALESCE(current_distance, '') as current_distance,
							COALESCE(potential_return, '') as potential_return,
							COALESCE(current_event_price, '') as current_event_price,
							COALESCE(current_event_position, '') as current_event_position,
							created_at,
							updated_at
						FROM EventPredictions
						WHERE event_date = ?  
							AND distanceTolerence < ? 
							AND average_position < ? 
							AND number_runs < ? 
						
						ORDER BY clean_bet_score DESC Limit 5`, params.EventDate, params.Delta, params.AvgPosition, params.TotalRuns)
	
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
			&racePrdiction.PotentialReturn,
			&racePrdiction.CurrentEventPrice,
			&racePrdiction.CurrentEventPosition,
			&racePrdiction.CreatedAt,
			&racePrdiction.UpdatedAt,
			
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		position, err := getPosition(racePrdiction.SelectionID, params.EventDate, db)

		if err != nil {
			racePrdiction.Position = "?"
		}else {
			racePrdiction.Position = position
			

		}	


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

		if p.CurrentEventPrice == "" {
			continue
		}

		odds_numerator := strings.Split(p.CurrentEventPrice, "/")[0]
		odds_denomenaor := strings.Split(p.CurrentEventPrice, "/")[1]

		// convert odds to float64
		oddsFloat_numberator, err := strconv.ParseFloat(odds_numerator, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		oddsFloat_denomenator, err := strconv.ParseFloat(odds_denomenaor, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		oddsFloat := oddsFloat_numberator/oddsFloat_denomenator
		if position == "1"{
			pnl = append(pnl, currentBet*oddsFloat + currentBet)
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

func formExit(formLastRunDate string, selectionId int, db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow(`
				Select count(*)
				from selectionsForm 
				where DATE(race_date) = ? and selection_id = ?;`, formLastRunDate, selectionId).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}