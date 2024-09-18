package preparation

import (
	"database/sql"
	"time"

	"net/http"

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

	rows, err := db.Query(`SELECT 	selection_id,
									selection_name,
									selection_link					
									FROM EventPredictions
									WHERE event_date = ?
									AND average_position < ? 
									AND number_runs < ?`,
									params.EventDate, params.AvgPosition, params.TotalRuns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var predictions []models.EventPrediction
	for rows.Next() {
		racePrdiction := models.EventPrediction{}
		err := rows.Scan(
			&racePrdiction.SelectionID,
			&racePrdiction.SelectionName,
			&racePrdiction.SelectionLink,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		predictions = append(predictions, racePrdiction)
	}


	err = SaveDateForm(predictions, c, params, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"simulationResults": predictions})
}
func SaveDateForm(predictions []models.EventPrediction, c *gin.Context, params models.GetWinnerParams, db *sql.DB) error {

	for _, prediction := range predictions {
		form, err := GetSelectionForm(prediction.SelectionLink)
		if err != nil {

			return err
		}

		for _, fr := range form {

			parsedEventDate, _ := time.Parse("2006-01-02", params.EventDate)
			if fr.EventDate == parsedEventDate {
				_, err = db.Exec(`UPDATE EventPredictions SET current_event_price = ?, current_event_position = ?  WHERE event_date = ? AND selection_id = ?`, fr.SPOdds, fr.Position, params.EventDate, prediction.SelectionID)
				if err != nil {
					return err
				}
			}
			exit, err := formExit(params.EventDate, prediction.SelectionID, db)
			if err != nil {
				return err
			}

			if !exit {
				err = SaveSelectionForm(db, fr, c, prediction.SelectionName, prediction.SelectionID)
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}

// Function to convert fractional odds to float64
// func fractionalOddsToFloat(odds string) (float64, error) {
// 	// Step 1: Remove any non-numeric characters except for the "/" using regex
// 	re := regexp.MustCompile(`[^\d/]+`)
// 	cleanedOdds := re.ReplaceAllString(odds, "")

// 	// Step 2: Split the odds by "/"
// 	parts := strings.Split(cleanedOdds, "/")
// 	if len(parts) != 2 {
// 		return 0, fmt.Errorf("invalid odds format: %s", odds)
// 	}

// 	// Step 3: Convert the numerator and denominator to float64
// 	numerator, err1 := strconv.ParseFloat(parts[0], 64)
// 	denominator, err2 := strconv.ParseFloat(parts[1], 64)

// 	if err1 != nil || err2 != nil {
// 		return 0, fmt.Errorf("error converting odds to float: %v, %v", err1, err2)
// 	}

// 	// Step 4: Return the decimal odds
// 	return numerator / denominator, nil
// }
