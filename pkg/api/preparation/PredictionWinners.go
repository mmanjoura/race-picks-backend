package preparation

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

// RacePicksSimulation handles the simulation of race picks and calculates win probabilities.
func GetPredictionWinners(c *gin.Context) {

	db := database.Database.DB

	params := models.GetWinnerParams{}

	// Bind JSON input to optimalParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stake, err := strconv.ParseFloat(params.Stake, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if stake == 0 {
		stake = 10.0
	}


	rows, err := db.Query(`SELECT id,
							event_link,
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
						WHERE event_date = ? 
							AND  distanceTolerence < ? 
							AND average_position < ? 
							AND number_runs < ? 
							AND current_distance BETWEEN 6 AND 14

						order by clean_bet_score DESC`, params.EventDate, params.Delta, params.AvgPosition, params.TotalRuns)
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
			&racePrdiction.EventLink,
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
		predictions = append(predictions, racePrdiction)
	}


	err = geLatestPrices(predictions, c, params, db, stake)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"simulationResults": predictions})
}
func geLatestPrices(predictions []models.EventPrediction, c *gin.Context, params models.GetWinnerParams, db *sql.DB, stake float64) error {

	for _, prediction := range predictions {

		predictionDetail, err := GetPredictionResult(prediction.SelectionName, prediction.EventLink)
		if err != nil {
			continue
		}
		result, err := fractionalOddsToFloat(predictionDetail.Price)
		if err != nil {
		 return err
		}
		potentialReturn := stake * result
		eventDate := prediction.EventDate[:10]

		_, err = db.Exec(`
							UPDATE EventPredictions SET 
										potential_return = ?,
										current_event_price = ?,
										current_event_position = ?
									WHERE DATE(event_date) = ? AND selection_id = ? `,
			potentialReturn,
			predictionDetail.Price,
			predictionDetail.Position,
			eventDate,
			prediction.SelectionID)

		if err != nil {
			return err
		}

	}
	return nil
}


// Function to convert fractional odds to float64
func fractionalOddsToFloat(odds string) (float64, error) {
    // Step 1: Remove any non-numeric characters except for the "/" using regex
    re := regexp.MustCompile(`[^\d/]+`)
    cleanedOdds := re.ReplaceAllString(odds, "")

    // Step 2: Split the odds by "/"
    parts := strings.Split(cleanedOdds, "/")
    if len(parts) != 2 {
        return 0, fmt.Errorf("invalid odds format: %s", odds)
    }

    // Step 3: Convert the numerator and denominator to float64
    numerator, err1 := strconv.ParseFloat(parts[0], 64)
    denominator, err2 := strconv.ParseFloat(parts[1], 64)

    if err1 != nil || err2 != nil {
        return 0, fmt.Errorf("error converting odds to float: %v, %v", err1, err2)
    }

    // Step 4: Return the decimal odds
    return numerator / denominator, nil
}

