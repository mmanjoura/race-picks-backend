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

func GetPredictions(c *gin.Context) {
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
			odds, 
			num_runners, 
			selection_position,
			bet_type,
			potential_return
		From EventPredictions where DATE(event_date) = ?;`, date)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer rows.Close()

	var eventName, totalScore, avgPosition, avgRating, eventDate,
		eventTime, selectionName, odds, numRunners, selectionPosition, betType, potentialReturn sql.NullString

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
			&numRunners,
			&selectionPosition,
			&betType,
			&potentialReturn,
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

		potentialReturnFloat, err := strconv.ParseFloat(nullableToString(potentialReturn), 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		selectionPostionInt := nullableToString(selectionPosition)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var price string
		if odds.Valid {
			price = removeDuplicateOdds(odds.String)
		} else {
			price = ""
		}
		racePrdiction.AvgRating = avgRatingFloat
		racePrdiction.EventName = nullableToString(eventName)
		racePrdiction.EventDate = nullableToString(eventDate)
		racePrdiction.EventTime = nullableToString(eventTime)
		racePrdiction.SelectionName = nullableToString(selectionName)
		racePrdiction.Odds = nullableToString(sql.NullString{String: price, Valid: true})
		racePrdiction.RunCount = nullableToString(numRunners)
		racePrdiction.SelectionPosition = selectionPostionInt
		racePrdiction.BetType = nullableToString(betType)
		racePrdiction.PotentialReturn = potentialReturnFloat

		racePrdictions = append(racePrdictions, racePrdiction)
	}



	// Return the meeting data
	c.JSON(http.StatusOK, gin.H{"predictions": racePrdictions})
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

// DetermineBetType determines the bet type based on the odds
func DetermineBetType(odds string) string {
	// Split the odds string by "/"
	parts := strings.Split(odds, "/")
	if len(parts) == 2 {
		numerator, _ := strconv.Atoi(parts[0])
		denominator, _ := strconv.Atoi(parts[1])

		// Check for BetType based on the odds
		if float64(numerator)/float64(denominator) < 1.0 {
			return "win bet"
		} else if float64(numerator)/float64(denominator) > 4.0 {
			return "place bet"
		}
	}
	// Default to an empty BetType if criteria are not met
	return ""
}

// CalculatePotentialReturn calculates the potential return based on BetType and odds
func CalculatePotentialReturn(betType string, odds string, amount float64) float64 {
	// Split the odds string by "/"
	parts := strings.Split(odds, "/")
	if len(parts) == 2 {
		numerator, _ := strconv.ParseFloat(parts[0], 64)
		denominator, _ := strconv.ParseFloat(parts[1], 64)

		// Calculate potential return for "win bet" or "place bet"
		if betType == "win bet" || betType == "place bet" {
			oddsMultiplier := (numerator / denominator) + 1
			return amount * oddsMultiplier
		}
	}
	// Default potential return is 0
	return 0
}



// Convert sql.NullFloat64 to a float64
func nullableToFloat(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0.0 // Return a default value (e.g., 0.0) if NULL
}
