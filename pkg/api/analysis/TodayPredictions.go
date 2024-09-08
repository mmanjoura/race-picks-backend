package analysis

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mmanjoura/race-picks-backend/pkg/api/common"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

// RacePicksSimulation handles the simulation of race picks and calculates win probabilities.
func GetTodayPredictions(c *gin.Context) {
	db := database.Database.DB
	var raceParams models.RaceParameters

	// Bind JSON input to optimalParams
	if err := c.ShouldBindJSON(&raceParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Query for today's runners
	rows, err := db.Query(`
		SELECT selection_id,
			selection_name,
			event_name,
			event_date,
			event_time,
			price,
			race_distance,
			race_category,
			track_condition,
			number_of_runners,
			race_track,
			race_class
		FROM EventRunners
		WHERE DATE(event_date) = ? `,
		raceParams.EventDate)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var selections []common.Selection
	for rows.Next() {
		var selection common.Selection

		// Use sql.NullString for nullable fields
		var selectionName, eventName, eventDate, eventTime, raceDistance, raceCategory, trackCondition, numberOfRunners, raceTrack, raceClass, odds sql.NullString

		// Scan the row values into the nullable types
		if err := rows.Scan(
			&selection.ID,
			&selectionName,
			&eventName,
			&eventDate,
			&eventTime,
			&odds,
			&raceDistance,
			&raceCategory,
			&trackCondition,
			&numberOfRunners,
			&raceTrack,
			&raceClass,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Convert sql.NullString to regular strings
		selection.Name = nullableToString(selectionName)
		selection.EventName = nullableToString(eventName)
		selection.EventDate = nullableToString(eventDate)
		selection.EventTime = nullableToString(eventTime)
		selection.RaceDistance = nullableToString(raceDistance)
		selection.RaceCategory = nullableToString(raceCategory)
		selection.TrackCondition = nullableToString(trackCondition)
		selection.NumberOfRunners = nullableToString(numberOfRunners)
		selection.RaceTrack = nullableToString(raceTrack)
		selection.RaceClass = nullableToString(raceClass)

		// Convert sql.NullFloat64 to float64 or set to a default value
		selection.Odds = nullableToString(odds)

		// Append the selection to the list
		selections = append(selections, selection)
	}

	// Get the selection with the least number of runs
	selectionCount, err := getSelectionWithLeastRuns(db, raceParams.EventName, raceParams.EventTime, raceParams.EventDate)
	_ = selectionCount // Ignore the result if not needed
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// var selectonsForm []models.SelectionForm
	var analysisData []models.AnalysisData

	for _, selection := range selections {
		// Execute the query
		rows, err = db.Query(`
			SELECT
				COALESCE(selection_id, 0),
				selection_name,	
				substr(position, 1, 1) as positon, 
				Age,
				Trainer,
				Sex,
				Sire,
				Dam,
				Owner,	
				race_class,					
				COUNT(*) AS num_runs,
				MAX(race_date) AS last_run_date,
				MAX(race_date) - MIN(race_date) AS duration,
				COUNT(CASE WHEN position = '1' THEN 1 END) AS win_count,
				AVG(position) AS avg_position,
				AVG(rating) AS avg_rating,
				AVG(distance) AS avg_distance_furlongs,
				AVG(sp_odds) AS sp_odds,
				GROUP_CONCAT(position, ', ') AS all_positions,
				GROUP_CONCAT(distance, ', ') AS all_distances,
				GROUP_CONCAT(racecourse, ', ') AS all_racecources,
				GROUP_CONCAT(DATE(race_date), ', ') AS all_race_dates 
			FROM
				SelectionsForm	WHERE selection_id = ?  order by race_date desc`, selection.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var data models.AnalysisData

		for rows.Next() {
			err := rows.Scan(
				&data.SelectionID,
				&data.SelectionName,
				&data.Position,
				&data.Age,
				&data.Trainer,
				&data.Sex,
				&data.Sire,
				&data.Dam,
				&data.Owner,
				&data.EventClass,
				&data.NumRuns,
				&data.LastRunDate,
				&data.Duration,
				&data.WinCount,
				&data.AvgPosition,
				&data.AvgRating,
				&data.AvgDistanceFurlongs,
				&data.AvgOdds,
				&data.AllPositions,
				&data.AllDistances,
				&data.AllCources,
				&data.AllRaceDates,
			)
			if err != nil {
				continue
			}
		}

		// Ignore selections with given parameters
		if yearExistsInDates(raceParams.Years, strings.Split(data.AllRaceDates, ",")) ||
			positionExistsInArray(raceParams.Positions, strings.Split(data.AllPositions, ",")) ||
			ageExistsInString(raceParams.Ages, data.Age) {

			continue
		}

		if data.SelectionID != 0 {

			analysisData = append(analysisData, data)
		}
	}

	mapResult := make(map[int]models.SelectionResult)
	var sortedResults []models.SelectionResult

	result := models.SelectionResult{}
	selectionsIds := []int{}

	// leastRuns := selectionCount[0].NumberOfRuns
	leastRuns := 1
	for _, data := range analysisData {
		if data.NumRuns < leastRuns {
			leastRuns = data.NumRuns
		}

		selectionsIds = append(selectionsIds, data.SelectionID)
	}

	newSelections := filterSelectionsByID(selections, selectionsIds)

	for id, selecion := range newSelections {

		if selecion.ID == analysisData[id].SelectionID {
			foatDistance := common.ParseDistance(selecion.RaceDistance)
			analysisData[id].CurrentDistance = foatDistance
			// if selecion.ID == 1148527 {
			// 	fmt.Print("Selection ID: ", selecion.ID)
			// }

			averagePostion := calculateAveragePosition(analysisData[id].AllPositions, leastRuns)
			totalScore := ScoreSelection(analysisData[id], raceParams, leastRuns)
			result.EventDate = selecion.EventDate
			result.SelectionID = selecion.ID
			result.EventName = selecion.EventName
			result.EventTime = selecion.EventTime
			result.SelectionName = selecion.Name
			result.Odds = selecion.Odds
			result.Trainer = analysisData[id].Trainer
			result.AvgPosition = math.Round(averagePostion)
			result.AvgRating = math.Round(analysisData[id].AvgRating)
			result.TotalScore = totalScore
			result.Age = analysisData[id].Age
			result.RunCount = analysisData[id].NumRuns
			mapResult[selecion.ID] = result

			sortedResults = append(sortedResults, result)
		}

	}

	// Step 2: Sort the slice by TotalScore
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].EventName > sortedResults[j].EventName
	})

	top3HighestScores := getTop3ScoresByTime(sortedResults)

	err = deletePredictions(db, raceParams.EventDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, result := range top3HighestScores {
		// bestSelection, err := FindBestSelection(result)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, r := range result {
			err = insertPredictions(db, r)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

	}
	c.JSON(http.StatusOK, gin.H{"simulationResults": top3HighestScores})
}

func insertPredictions(db *sql.DB, data models.SelectionResult) error {

	// Prepare the INSERT statement
	stmt, err := db.Prepare(`
		INSERT INTO EventPredictions (event_date, selection_id, selection_name, odds, clean_bet_score, average_position, average_rating, event_name, event_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the INSERT statement
	_, err = stmt.Exec(data.EventDate, data.SelectionID, data.SelectionName, data.Odds, data.TotalScore, data.AvgPosition, data.AvgRating, data.EventName, data.EventTime)
	if err != nil {
		return err
	}

	// Return nil if no error occurred
	return nil
}
func deletePredictions(db *sql.DB, eventDate string) error {
	// Check if a record with the same event_date and selection_id exists
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM EventPredictions
			WHERE DATE(event_date) = ?
		)
	`, eventDate).Scan(&exists)
	if err != nil {
		return err
	}

	// If a record exists, delete it
	if exists {
		_, err = db.Exec(`
			DELETE FROM EventPredictions
			WHERE DATE(event_date) = ?
		`, eventDate)
		if err != nil {
			return err
		}
	}

	// Return nil if no error occurred
	return nil
}

// FindBestSelection returns the selection with the highest score, highest rating, and youngest age
func FindBestSelection(data []models.SelectionResult) (models.SelectionResult, error) {
	var bestSelection models.SelectionResult
	var found bool

	// Iterate through the map to find the best selection

	for _, record := range data {
		// Extract age from the name (e.g., "Little Empire 4" -> 4)
		age := extractAge(record.SelectionName)

		// Compare based on the criteria: highest score, highest rating, and youngest age
		if !found || isBetterSelection(record, bestSelection, age) {
			bestSelection = record
			bestSelection.Age = strconv.Itoa(age) // Convert age to string and assign it
			found = true
		}
	}

	if !found {
		return models.SelectionResult{}, fmt.Errorf("no selections found")
	}

	return bestSelection, nil
}

// Extracts the age from the name of the selection
func extractAge(name string) int {
	// Split the name by space and get the last part, assuming it's the age
	parts := strings.Split(name, " ")
	age := 0
	if len(parts) > 0 {
		fmt.Sscanf(parts[len(parts)-1], "%d", &age) // Read age as integer
	}
	return age
}

// Checks if the current selection is better based on the criteria
func isBetterSelection(current, best models.SelectionResult, currentAge int) bool {
	bestAge, _ := strconv.Atoi(best.Age) // Convert best.Age from string to int
	if current.TotalScore > best.TotalScore {
		return true
	} else if current.TotalScore == best.TotalScore {
		if current.AvgRating > best.AvgRating {
			return true
		} else if current.AvgRating == best.AvgRating {
			if currentAge < bestAge { // Compare currentAge with bestAge
				return true
			}
		}
	}
	return false
}
