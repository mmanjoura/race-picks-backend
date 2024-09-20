package preparation

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
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

	params.Stake = 10.0

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

func saveWinners(db *sql.DB, winner models.Winner) error {

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback() // Rollback the transaction in case of error
		} else {
			tx.Commit() // Commit the transaction if all goes well
		}
	}()

	_, err = tx.Exec(`DELETE FROM Winners WHERE selection_id = ?`, winner.SelectionID)
	if err != nil {
		log.Println("Failed to delete existing records:", err)
		return err
	}

	// Step 2: INSERT new record
	_, err = tx.Exec(`
    INSERT INTO Winners (
        selection_id, 
        selection_name, 
        current_odds, 
        current_position, 
        current_return,
		event_date
    ) 
    VALUES (?, ?, ?, ?, ?, ?);`,

		winner.SelectionID,
		winner.SelectionName,
		winner.CurrentOdds,
		winner.CurrentPosition,
		winner.CurrentReturn,
		winner.EventDate)
	if err != nil {
		return err
	}
	return nil
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

				currentPostion := strings.Split(fr.Position, "/")[0]
				if currentPostion == "1" {

					nom := strings.Split(fr.SPOdds, "/")[0]
					denom := strings.Split(fr.SPOdds, "/")[1]

					nomFloatValue, err := strconv.ParseFloat(nom, 64)
					if err != nil {
						return err
					}

					denomFloatValue, err := strconv.ParseFloat(denom, 64)
					if err != nil {
						return err
					}

					var winner models.Winner
					floatStake := params.Stake
					winner.SelectionID = prediction.SelectionID
					winner.SelectionName = prediction.SelectionName
					winner.CurrentOdds = nomFloatValue / denomFloatValue
					winner.CurrentPosition = fr.Position
					winner.CurrentReturn = (winner.CurrentOdds * floatStake) + floatStake
					winner.EventDate = params.EventDate

					err = saveWinners(db, winner)
					if err != nil {
						return err
					}

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
