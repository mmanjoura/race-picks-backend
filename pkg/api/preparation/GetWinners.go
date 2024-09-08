package preparation

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

func MeetingWinners(c *gin.Context) {
	db := database.Database.DB

	var racePrdictions []models.SelectionResult



	// Query for today's runners
	date := c.Query("event_date")

	rows, err := db.Query(`
		SELECT selection_id,
			clean_bet_score,
			average_position,
			average_rating,
			event_name,
			event_date,
			event_time,
			selection_name,
			odds
		From EventPredictions where DATE(event_date) = ?;`, date)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer rows.Close()

	var eventName, totalScore, avgPosition, avgRating, eventDate, eventTime, selectionName, odds sql.NullString

	for rows.Next() {
		racePrdiction := models.SelectionResult{}
		err := rows.Scan(
			&racePrdiction.SelectionID,
			&totalScore,
			&avgPosition,
			&avgRating,
			&eventName,
			&eventDate,
			&eventTime,
			&selectionName,
			&odds,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		totalScoreFloat, err := strconv.ParseFloat(nullableToString(totalScore), 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		racePrdiction.TotalScore = totalScoreFloat
		avgPositionFloat, err := strconv.ParseFloat(nullableToString(avgPosition), 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		racePrdiction.AvgPosition = avgPositionFloat
		avgRatingFloat, err := strconv.ParseFloat(nullableToString(avgRating), 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		racePrdiction.AvgRating = avgRatingFloat
		racePrdiction.EventName = nullableToString(eventName)
		racePrdiction.EventDate = nullableToString(eventDate)
		racePrdiction.EventTime = nullableToString(eventTime)
		racePrdiction.SelectionName = nullableToString(selectionName)
		racePrdiction.Odds = nullableToString(odds)

		racePrdictions = append(racePrdictions, racePrdiction)
	}

	var todayBets []models.SelectionResult
	var selectionTime string
	for i := 0; i < len(racePrdictions); i++ {

		odds, err := calculateOdds(racePrdictions[i].Odds)
		if err != nil {
			if err.Error() == "Invalid input" {
				continue
			}
		}

		// we only want to bet on selections with odds between 10 and 20
		if odds < 20 && odds > 10 {
			if racePrdictions[i].EventTime == selectionTime {
				continue
			}
			todayBets = append(todayBets, racePrdictions[i])
			selectionTime = racePrdictions[i].EventTime
		}
	}

	// Return the meeting data
	c.JSON(http.StatusOK, gin.H{"predictions": todayBets})
}

func calculateOdds(input string) (float64, error) {

	// Split the input string by "/"
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return 0, errors.New("Invalid input")
	}

	// Convert the string parts to float64 numbers
	numerator, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}

	denominator, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}

	// Perform the division
	if denominator == 0 {
		return 0, errors.New("Division by zero")
	}

	result := numerator / denominator

	return result, nil
}
