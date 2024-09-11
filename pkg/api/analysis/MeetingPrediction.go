package analysis

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mmanjoura/race-picks-backend/pkg/api/common"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

// RacePicksSimulation handles the simulation of race picks and calculates win probabilities.
func GetMeetingPrediction(c *gin.Context) {
	db := database.Database.DB
	var raceParams models.RaceParameters

	// Bind JSON input to optimalParams
	if err := c.ShouldBindJSON(&raceParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// betAmount, err := strconv.ParseFloat(raceParams.BetAmount, 64)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

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
			race_class,
			selection_link,
			event_link
		FROM EventRunners
		WHERE DATE(event_date) = ?  AND event_name = ? AND event_time = ?`,
		raceParams.EventDate, raceParams.EventName, raceParams.EventTime)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var selections []common.Selection
	for rows.Next() {
		var selection common.Selection

		// Use sql.NullString for nullable fields
		var selectionName, eventName, eventDate, eventTime, raceDistance, raceCategory, trackCondition, numberOfRunners, raceTrack, raceClass, odds, eventLink sql.NullString

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
			&selection.Link,
			&eventLink,
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
		selection.EventLink = nullableToString(eventLink)

		// Convert sql.NullFloat64 to float64 or set to a default value
		selection.Odds = nullableToString(odds)

		// Append the selection to the list
		selections = append(selections, selection)
	}

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
				position, 
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
		data.EventLink = selection.EventLink

		currentDistance := convertDistance(selection.RaceDistance)
		distance, err := strconv.ParseFloat(currentDistance, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		data.CurrentDistance = distance

		if selection.ID == 1154889 {
			fmt.Println("Selection ID: ", selection.ID)
		}

		// Ignore selections with given parameters
		if yearExistsInDates(raceParams.Years, strings.Split(data.AllRaceDates, ",")) ||
			positionExistsInArray(raceParams.Positions, strings.Split(data.AllPositions, ",")) ||
			ageExistsInString(raceParams.Ages, data.Age) {

			continue
		}

		// Add filter here to ignre what you dont want

		if data.SelectionID != 0 {
			data.NumberOfRunners = selection.NumberOfRunners
			data.SelecionLink = selection.Link

			ageString := strings.Split(data.Age, " ")[0]
			age, err := strconv.Atoi(ageString)
			if err != nil {
				fmt.Println("Error converting to integer:", err)
			}
			// Age limit
			if age > 8 {
				continue
			}

			if strings.Contains(data.EventLink, "handicap") {
				continue
			}

			analysisData = append(analysisData, data)
		}

	}

	mapResult := make(map[int]models.SelectionResult)
	var sortedResults []models.SelectionResult

	result := models.SelectionResult{}
	var leastRuns int
	selectionsIds := []int{}

	for _, data := range analysisData {
		dates := strings.Split(data.AllRaceDates, ",")
		leastRuns = len(dates)
		if leastRuns < len(dates) {
			leastRuns = len(dates)
		}

		perferedDistancd := preferredDistance(data.AllPositions, data.AllDistances, data.AllRaceDates)
		if strings.Contains(perferedDistancd, "f") || strings.Contains(perferedDistancd, "m") {
			// convert to furlong
			perferedDistancd = convertDistance(perferedDistancd)
		}

		foatDistance, err := strconv.ParseFloat(perferedDistancd, 64)
		if err != nil {
			continue
		}

		diff := math.Abs(foatDistance - data.CurrentDistance)
		if diff > 1 {
			continue
		}

		selectionsIds = append(selectionsIds, data.SelectionID)
	}

	newSelections := filterSelectionsByID(selections, selectionsIds)

	for id, selecion := range newSelections {
		// if leastRuns < 4 continue
		if leastRuns < 4 {
			continue
		}

		NumRunnder := strings.Split(selecion.NumberOfRunners, " ")[0]

		numerOfRunner, err := strconv.Atoi(NumRunnder)

		if err != nil {
			continue
		}

		if numerOfRunner > 10 {
			continue
		}

		NumberOfrunners, err := extractNumber(analysisData[id].NumberOfRunners)
		if err != nil {
			continue
		}
		if selecion.ID == analysisData[id].SelectionID {

			foatDistance := common.ParseDistance(selecion.RaceDistance)
			analysisData[id].CurrentDistance = foatDistance

			averagePostion := calculateAveragePosition(analysisData[id].AllPositions, leastRuns)
			if averagePostion < 4 {
				continue
			}
			analysisData[id].AvgPosition = averagePostion
			totalScore := ScoreSelection(analysisData[id], raceParams, leastRuns)
			result.EventName = selecion.EventName
			result.EventTime = selecion.EventTime
			result.SelectionName = selecion.Name
			result.Odds = selecion.Odds
			result.Trainer = analysisData[id].Trainer
			result.AvgRating = math.Round(analysisData[id].AvgRating)
			result.TotalScore = totalScore
			result.Age = analysisData[id].Age
			result.SelectionID = selecion.ID
			result.EventDate = raceParams.EventDate
			result.RunCount = NumberOfrunners
			result.SelectionLink = selecion.Link
			result.SelectionPosition = analysisData[id].Position

			mapResult[selecion.ID] = result

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			sortedResults = append(sortedResults, result)
		}

	}

	// Step 2: Sort the slice by TotalScore
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].EventName > sortedResults[j].EventName
	})

	// convert sortedResults to map of sting and []SelectionResult
	mpResult := make(map[string][]models.SelectionResult)
	for _, result := range sortedResults {
		// remove any score less than 300
		if result.TotalScore < 300 {
			continue

		}

		if strings.Contains(result.Odds, "/") {
			result.Odds = strings.Split(result.Odds, "/")[0]

			// convert postion into int
			// odds, err := strconv.Atoi(result.Odds)
			if err != nil {
				continue
			}
			// if odds > 5 {
			// 	continue
			// }

		}

		mpResult[result.EventTime] = append(mpResult[result.EventTime], result)

		if len(mapResult) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"Not found error": err.Error()})
			return
		}

	}

	// top3HighestScores := getTop3ScoresByTime(sortedResults)
	// top2HighestScores := getTop2ScoresByTime(sortedResults)
	// top1HighestScores := getTopScoreByTime(sortedResults)
	// if strings.Contains(selections[i].SelectionPosition, "/") {
	// 	selections[i].SelectionPosition = strings.Split(selections[i].SelectionPosition, "/")[0]
	// }

	// do not pay attention, this mean only run prediction for entire day instead of single event.
	if raceParams.Going == "Good" {
		// resultWithPrices := AddBetTypeAndReturnsToSelections(mpResult, betAmount, raceParams.EventDate) // Assuming a bet amount of 100
		if len(mpResult) != 0 {
			err = deletePredictions(db, raceParams.EventDate, raceParams.EventName, raceParams.EventTime)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			for _, result := range mpResult {

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
		}
	}

	c.JSON(http.StatusOK, gin.H{"simulationResults": mpResult})
}

func filterSelectionsByID(selections []common.Selection, ids []int) []common.Selection {
	idSet := make(map[int]struct{})
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	var filteredSelections []common.Selection
	for _, selection := range selections {
		if _, exists := idSet[selection.ID]; exists {
			filteredSelections = append(filteredSelections, selection)
		}
	}

	return filteredSelections
}

func fetchConstantScore(db *sql.DB, category, item string) (float64, error) {
	var score float64
	row := db.QueryRow("SELECT score FROM score_constants WHERE category = ? AND item = ?", category, item)
	err := row.Scan(&score)
	if err != nil {
		return 0.0, err
	}
	return score, nil
}
func scoreRace(db *sql.DB, races []models.AnalysisData, raceParams models.RaceParameters) (models.ScoreBreakdown, error) {
	var raceTypeScore, courseScore, distanceScore, classScore, ageScore, positionScore float64
	var eventName, eventTime, selectionName, odds, trainer string

	return models.ScoreBreakdown{

		EventName:     eventName,
		EventTime:     eventTime,
		SelectionName: selectionName,
		Odds:          odds,
		Trainer:       trainer,
		RaceTypeScore: raceTypeScore,
		CourseScore:   courseScore,
		DistanceScore: distanceScore,
		ClassScore:    classScore,
		AgeScore:      ageScore,
		PositionScore: positionScore,
	}, nil
}

// CheckImprovement checks if the horse is improving over the distance.
func CheckImprovement(data []models.HistoricalData) string {
	sort.Slice(data, func(i, j int) bool {
		return data[i].Date.Before(data[j].Date)
	})

	improving := true
	for i := 1; i < len(data); i++ {
		if data[i].Distance < data[i-1].Distance {
			improving = false
			break
		}
	}

	if improving {
		return "Good Score"
	}
	return "Bad Score"
}

// ParseHistoricalData parses the historical data from a slice of strings.
func ParseHistoricalData(data [][]string) ([]models.HistoricalData, error) {
	var historicalData []models.HistoricalData
	for _, row := range data {

		date, err := time.Parse("2006-01-02", row[0])
		if err != nil {
			return nil, err
		}
		strDisting := convertDistance(row[1])
		distance, err := strconv.ParseFloat(strDisting, 64)
		if err != nil {
			return nil, err
		}
		historicalData = append(historicalData, models.HistoricalData{
			Date:     date,
			Position: row[0],
			Distance: distance,
		})
	}
	return historicalData, nil
}

func convertDistance(distanceStr string) string {
	// if this string contain "."
	if strings.Contains(distanceStr, ".") {
		alreadyFormated := strings.Split(distanceStr, ".")
		if len(alreadyFormated[0]) > 0 {
			return distanceStr
		}
	}

	_, err := strconv.ParseFloat(distanceStr, 64)
	if err == nil {
		return distanceStr
	}

	parts := strings.Split(distanceStr, " ")
	furlongs := 0.0
	for _, part := range parts {
		if strings.Contains(part, "m") {
			miles, err := strconv.ParseFloat(strings.TrimSuffix(part, "m"), 64)
			if err == nil {
				furlongs += miles * 8
			}
		} else if strings.Contains(part, "f") {
			f, err := strconv.ParseFloat(strings.TrimSuffix(part, "f"), 64)
			if err == nil {
				furlongs += f
			}
		} else if strings.Contains(part, "y") {
			// Assume 220 yards = 1 furlong (approximately)
			yards, err := strconv.ParseFloat(strings.TrimSuffix(part, "y"), 64)
			if err == nil {
				furlongs += yards / 220.0
			}
		}
	}
	return strconv.FormatFloat(furlongs, 'f', -1, 64)
}

// Helper function to parse race position
func parsePosition(pos string) (position int, fieldSize int) {
	parts := strings.Split(pos, "/")
	position, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0
	}

	return position, fieldSize
}

// Helper function to parse rating
func parseRating(rating string) float64 {
	if rating == "-" {
		return 0 // Or any neutral value for missing ratings
	}
	parsedRating, _ := strconv.ParseFloat(rating, 64)
	return parsedRating
}

// Helper function to calculate date score based on recency
func calculateDateScore(runDate string) float64 {
	const layout = "2006-01-02"
	date, _ := time.Parse(layout, runDate)
	daysAgo := time.Since(date).Hours() / 24.0

	// Closer races get higher score
	if daysAgo <= 30 {
		return 10.0
	} else if daysAgo <= 60 {
		return 8.0
	} else if daysAgo <= 90 {
		return 6.0
	} else {
		return 4.0
	}
}

// Main function to calculate scores
func calculateHorseScores(selectionRunDates, selectionPositions, selectionsRating []string) models.ScoreBreakdown {
	var dateScore, positionScore, ratingScore float64

	// Iterate through all races
	for i := range selectionRunDates {
		// Date Score
		dateScore += calculateDateScore(selectionRunDates[i])

		// Position Score
		position, fieldSize := parsePosition(selectionPositions[i])
		if fieldSize == 0 {
			continue
		}
		positionScore += 10.0 - float64(position)/float64(fieldSize)

		// Rating Score
		ratingScore += parseRating(selectionsRating[i])
	}
	return models.ScoreBreakdown{

		DateScore:     dateScore,
		PositionScore: positionScore,
		RatingScore:   ratingScore,
	}
}

// New function to fetch age score based on the race distance
func fetchAgeScore(db *sql.DB, age int, distance float64) (float64, error) {
	var score float64
	var err error

	if distance > 12.0 {
		// Greater than 12 furlongs
		score, err = fetchConstantScore(db, "Age-greater-12f", strconv.Itoa(age))
	} else {
		// Less than or equal to 12 furlongs
		score, err = fetchConstantScore(db, "Age-bellow-12f", strconv.Itoa(age))
	}

	return score, err
}

// Function to filter SelectionForms based on selection_id
func filterSelectionFormsByID(forms []models.SelectionForm, selectionID int) []models.SelectionForm {
	var filteredForms []models.SelectionForm

	for _, form := range forms {
		if form.SelectionID == selectionID {
			filteredForms = append(filteredForms, form)
		}
	}

	return filteredForms
}

func stringInSlice(target string, slice []string) bool {
	for _, item := range slice {
		item := strings.Split(item, "-")[0]
		if item == target {
			return true
		}
	}
	return false
}

// Scoring Function
func ScoreSelection(selection models.AnalysisData, params models.RaceParameters, limit int) float64 {

	var score float64

	// We only want to deal with horses that have run at least 3 times

	// 0. gelding Score
	if selection.Sex == "Gelding" {
		score += 5
	}

	// 1. Number of years in competions
	if selection.Duration < 4 {
		score += 5
	}

	// 2. Age Score
	ageString := strings.Split(selection.Age, " ")[0]
	age, err := strconv.Atoi(ageString)
	if err != nil {
		fmt.Println("Error converting to integer:", err)
	}
	if selection.CurrentDistance < 12.0 {
		if age <= 7.0 {
			score += 5
		}
	} else {
		if age >= 4.0 && age <= 8.0 {
			score += 5
		}
	}

	// 3. Number of runs so far
	if selection.NumRuns < 10 {
		score += 5
	}

	// 5. Rating
	if selection.AvgRating > 0 && selection.AvgRating < 10 {
		score += 2.5
	}
	if selection.AvgRating > 10 && selection.AvgRating < 20 {
		score += 5
	}
	if selection.AvgRating > 20 && selection.AvgRating < 40 {
		score += 7.5
	}
	if selection.AvgRating > 40 {
		score += 10
	}

	// 6. Odds
	if selection.AvgOdds < 10 {
		score += 5
	}

	// Distance Analysis
	distances := strings.Split(selection.AllDistances, ",")
	if len(distances) > limit {
		distances = distances[:limit]
	}

	var totalDistance float64
	for _, distance := range distances {
		distance = strings.TrimSpace(distance)
		convertedDistance := convertDistance(distance)
		fd, err := strconv.ParseFloat(convertedDistance, 64)
		if err != nil {
			continue
		}
		totalDistance += fd
	}

	avgDistance := totalDistance / float64(len(distances))

	// Distance and Age-Based Scoring
	distanceDiff := math.Abs(avgDistance - selection.CurrentDistance)
	switch {
	case avgDistance <= 12:

		score += calculateDistanceScore(distanceDiff, []float64{0.5, 1.0, 1.5}, []float64{30, 15, 10, 8, 5})
	default:

		score += calculateDistanceScore(distanceDiff, []float64{1.5, 2.0, 3.5}, []float64{30, 15, 10, 8, 5})
	}

	// Position Analysis
	positions := strings.Split(selection.AllPositions, ",")
	score += calculatePositionScore(positions, limit)

	return score
}

func safeDivide(numerator, denominator float64) float64 {
	if denominator == 0.0 {
		// Handle division by zero case, maybe return 0 or some other default value.
		fmt.Println("Warning: Division by zero detected.")
		return math.Inf(1) // Or return a different value that makes sense in your context.
	}
	return numerator / denominator
}

// Function to get the highest score selection for each event time
func getHighestScoreByTime(sortedResults []models.SelectionResult) map[string]models.SelectionResult {
	// Group the results by EventTime
	groupedResults := make(map[string]models.SelectionResult)

	for _, result := range sortedResults {
		currentBest, exists := groupedResults[result.EventTime]
		if !exists || result.TotalScore > currentBest.TotalScore {
			groupedResults[result.EventTime] = result
		}
	}

	return groupedResults
}

// Function to get the second highest score selection for each event time
func getSecondHighestScoreByTime(sortedResults []models.SelectionResult) map[string]models.SelectionResult {
	// Group the results by EventTime
	groupedResults := make(map[string][]models.SelectionResult)

	for _, result := range sortedResults {
		groupedResults[result.EventTime] = append(groupedResults[result.EventTime], result)
	}

	// Prepare the map to store the second highest score selections
	secondHighestResults := make(map[string]models.SelectionResult)

	for time, results := range groupedResults {
		// Sort the results by TotalScore in descending order
		sort.Slice(results, func(i, j int) bool {
			return results[i].TotalScore > results[j].TotalScore
		})

		// Check if there are at least two results to get the second highest
		if len(results) > 1 {
			secondHighestResults[time] = results[1] // Index 1 is the second highest
		}
	}

	return secondHighestResults
}

// Function to check if any year from the input string exists in the dates array
func yearExistsInDates(yearsStr string, dates []string) bool {
	// Split the years string by comma and trim spaces
	years := strings.Split(yearsStr, ",")
	for i, year := range years {
		years[i] = strings.TrimSpace(year)
	}

	// Iterate over the dates array to check for each year
	for _, date := range dates {
		dateYear := strings.Split(date, "-")[0] // Extract year from the date

		// Check if extracted year exists in the list of years
		for _, year := range years {
			if dateYear == year {
				return true
			}
		}
	}

	return false
}

// Function to check if any position from the input string exists in the positions array
func positionExistsInArray(positionsStr string, positionsArray []string) bool {
	// Split the positions string by comma and trim spaces
	positions := strings.Split(positionsStr, ",")
	for i, pos := range positions {
		positions[i] = strings.TrimSpace(pos)
	}

	// Iterate over the positionsArray to check for each position
	for _, pos := range positionsArray {
		// Split the position by '/' and extract the second part
		parts := strings.Split(pos, "/")
		if len(parts) < 2 {
			continue // Skip if the format is invalid
		}
		horsePosition := parts[1]

		// Check if the extracted horse position exists in the list of positions
		for _, position := range positions {
			if horsePosition == position {
				return true
			}
		}
	}

	return false
}

// Function to check if the age exists in the given string of ages
func ageExistsInString(agesStr string, ageVariable string) bool {
	// Split the ages string by comma and trim spaces
	ages := strings.Split(agesStr, ",")
	for i, age := range ages {
		ages[i] = strings.TrimSpace(age)
	}

	// Use a regular expression to extract the age from the age variable string
	re := regexp.MustCompile(`\d+`)
	matchedAge := re.FindString(ageVariable)

	// Check if the extracted age exists in the list of ages
	for _, age := range ages {
		if matchedAge == age {
			return true
		}
	}

	return false
}

func getTop3ScoresByTime(sortedResults []models.SelectionResult) map[string][]models.SelectionResult {
	// Group the results by EventTime
	groupedResults := make(map[string][]models.SelectionResult)

	for _, result := range sortedResults {
		// Get the current list of results for the EventTime
		currentResults, exists := groupedResults[result.EventTime]

		if !exists {
			// If there are no results yet for this EventTime, add the current result
			groupedResults[result.EventTime] = []models.SelectionResult{result}
		} else {
			// Append the current result to the existing list
			currentResults = append(currentResults, result)

			// Sort the results by TotalScore in descending order
			sort.Slice(currentResults, func(i, j int) bool {
				return currentResults[i].TotalScore > currentResults[j].TotalScore
			})

			// Keep only the top 3 scores
			if len(currentResults) > 3 {
				currentResults = currentResults[:3]
			}

			groupedResults[result.EventTime] = currentResults
		}
	}

	return groupedResults
}

func calculateAveragePosition(positionsString string, n int) float64 {
	// Split the input string into an array of positions
	positionsArray := strings.Split(positionsString, ", ")

	// Check if n is greater than the length of the positions array
	if n > len(positionsArray) {
		n = len(positionsArray)
	}

	// Variable to store the total of the numerators
	totalPosition := 0

	// Iterate over the first n elements to calculate the sum of positions
	for i := 0; i < n; i++ {
		// Split each position into numerator and denominator
		positionParts := strings.Split(positionsArray[i], "/")

		// Convert the numerator to an integer
		numerator, err := strconv.Atoi(positionParts[0])
		if err != nil {
			fmt.Println("Error converting to integer:", err)
			continue
		}

		// Add the numerator to the total
		totalPosition += numerator
	}

	// Calculate the average position
	averagePosition := float64(totalPosition) / float64(n)
	return averagePosition
}

// Function to check if the age exists in the given string of ages
func ageInString(agesStr string) string {
	// Split the ages string by comma and trim spaces
	ages := strings.Split(agesStr, ",")
	for i, age := range ages {
		ages[i] = strings.TrimSpace(age)
	}

	// Use a regular expression to extract the age from the age variable string
	re := regexp.MustCompile(`\d+`)
	matchedAge := re.FindString(agesStr)

	return matchedAge
}

// Convert sql.NullFloat64 to a float64
func nullableToFloat(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0.0 // Return a default value (e.g., 0.0) if NULL
}

// Convert sql.NullString to a regular string
func nullableToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "" // Return empty string if NULL
}

// Calculate Distance Score
func calculateDistanceScore(distanceDiff float64, thresholds, scores []float64) float64 {
	var score float64
	for i, threshold := range thresholds {
		if distanceDiff <= threshold {
			score += scores[i]
			break
		}
	}
	return score
}

// Calculate Position Score
func calculatePositionScore(positions []string, limit int) float64 {
	var score float64
	if len(positions) > limit {
		positions = positions[:limit]
	}

	for _, pos := range positions {
		pos = strings.TrimSpace(pos)
		if strings.Contains(pos, "F") || strings.Contains(pos, "PU") || strings.Contains(pos, "U") || strings.Contains(pos, "R") {
			score -= 5
		}
		if strings.Contains(pos, "/") {
			p := strings.Split(pos, "/")
			if len(p) != 2 {
				continue
			}
			numerator, err1 := strconv.Atoi(strings.TrimSpace(p[0]))
			denominator, err2 := strconv.Atoi(strings.TrimSpace(p[1]))
			if err1 != nil || err2 != nil || denominator == 0 {
				score -= 1
				continue
			}
			score += math.Round(safeDivide(float64(denominator), float64(numerator))) * 10
		}
	}
	return score
}

func extractNumber(s string) (string, error) {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return "", fmt.Errorf("no number found")
	}
	number, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", err
	}
	return strconv.Itoa(number), nil
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
		if betType == "win bet" {
			oddsMultiplier := (numerator / denominator)
			return amount * oddsMultiplier
		}
		if betType == "place bet" {
			oddsMultiplier := (numerator / denominator)
			return amount * oddsMultiplier
		}

	}
	// Default potential return is 0
	return 0
}

// AddBetTypeAndReturnsToSelections processes the input map and adds BetType, SelectionPosition, and PotentialReturn fields
func AddBetTypeAndReturnsToSelections(selectionsMap map[string][]models.SelectionResult, amount float64, date string) map[string][]models.SelectionResult {
	for eventTime, selections := range selectionsMap {
		for i := range selections {
			// Determine the BetType for each selection
			selections[i].BetType = DetermineBetType(selections[i].Odds)

			if strings.Contains(selections[i].SelectionPosition, "/") {
				selections[i].SelectionPosition = strings.Split(selections[i].SelectionPosition, "/")[0]
			}

			// Calculate the PotentialReturn only if SelectionPosition is 1
			if selections[i].SelectionPosition == "1" {
				selections[i].PotentialReturn = CalculatePotentialReturn(selections[i].BetType, selections[i].Odds, amount)
				// } else if selections[i].SelectionPosition == "2" {
				// 	selections[i].PotentialReturn = 0.25 * CalculatePotentialReturn(selections[i].BetType, selections[i].Odds, amount)
				// } else if selections[i].SelectionPosition == "3" {
				// 	selections[i].PotentialReturn = 0.25 * CalculatePotentialReturn(selections[i].BetType, selections[i].Odds, amount)
			} else {
				selections[i].PotentialReturn = 0 // No potential return if position is not 1
			}

		}

		// Update the selections in the map
		selectionsMap[eventTime] = selections
	}
	return selectionsMap
}

// Function to get the selection with the least number of runs
func getSelectionWithLeastRuns(db *sql.DB, eventName, eventTime, eventDate string) ([]Selection, error) {
	// SQL query to find the selection ID with the least number of runs

	var selections []Selection
	rows, err := db.Query(`
					SELECT 
						SelectionsForm.selection_id,
						EventRunners.selection_name,
						EventRunners.event_date,
						COUNT(*) AS number_of_runs
					FROM 
						SelectionsForm 
						INNER JOIN EventRunners ON SelectionsForm.selection_id = EventRunners.selection_id
						WHERE SelectionsForm.racecourse = ? and EventRunners.event_time = ? and DATE(EventRunners.event_date) = ?
					GROUP BY 
						SelectionsForm.selection_id
						Order by number_of_runs`, eventName, eventTime, eventDate)

	if err != nil {
		return []Selection{}, err
	}
	defer rows.Close()

	var selection Selection

	// Get the result
	if rows.Next() {
		err := rows.Scan(
			&selection.ID,
			&selection.Name,
			&selection.EventDate,
			&selection.NumberOfRuns,
		)
		if err != nil {
			return []Selection{}, err
		}
		selections = append(selections, selection)
	}

	// Check for any error encountered during iteration
	if err := rows.Err(); err != nil {
		return []Selection{}, err
	}

	// Return the selection ID and number of runs
	return selections, nil
}

func getTopScoreByTime(sortedResults []models.SelectionResult) map[string][]models.SelectionResult {
	// Create a map to store the top score by EventTime
	topScores := make(map[string][]models.SelectionResult)

	for _, result := range sortedResults {
		// Check if there's already a result for this EventTime
		currentTop, exists := topScores[result.EventTime]

		if !exists || result.TotalScore > currentTop[0].TotalScore {
			// If there's no result yet or the current result has a higher score, update it
			topScores[result.EventTime] = []models.SelectionResult{result}
		}
	}

	return topScores
}

func getTop2ScoresByTime(sortedResults []models.SelectionResult) map[string][]models.SelectionResult {
	// Create a map to store the top 2 scores by EventTime
	top2Scores := make(map[string][]models.SelectionResult)

	for _, result := range sortedResults {
		// Check if there are already results for this EventTime
		currentTopScores, exists := top2Scores[result.EventTime]

		if !exists {
			// If no results yet for this EventTime, initialize with the current result
			top2Scores[result.EventTime] = []models.SelectionResult{result}
		} else {
			// Append the current result to the existing list
			currentTopScores = append(currentTopScores, result)

			// Sort the results by TotalScore in descending order
			sort.Slice(currentTopScores, func(i, j int) bool {
				return currentTopScores[i].TotalScore > currentTopScores[j].TotalScore
			})

			// Keep only the top 2 scores
			if len(currentTopScores) > 2 {
				currentTopScores = currentTopScores[:2]
			}

			// Update the map with the top 2 results
			top2Scores[result.EventTime] = currentTopScores
		}
	}

	return top2Scores
}

func preferredDistance(performances, distances, dates string) string {
	// Split input strings into slices
	performanceList := strings.Split(performances, ", ")
	distanceList := strings.Split(distances, ", ")
	dateList := strings.Split(dates, ", ")

	// Check if the lengths match
	if len(performanceList) != len(distanceList) || len(distanceList) != len(dateList) {
		return "Data length mismatch"
	}

	// Create a map to store total scores and count per distance
	distanceScores := make(map[string]float64)
	distanceCount := make(map[string]int)

	// Calculate scores based on performance
	for i := 0; i < len(performanceList); i++ {
		parts := strings.Split(performanceList[i], "/")
		if len(parts) != 2 {
			continue
		}

		// Parse position and total runners
		position, err1 := strconv.Atoi(parts[0])
		totalRunners, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			continue
		}

		// Calculate score (lower is better)
		score := float64(position) / float64(totalRunners)

		// Update the total score and count for the distance
		distanceScores[distanceList[i]] += score
		distanceCount[distanceList[i]]++
	}

	// Determine the preferred distance with the lowest average score
	var preferredDistance string
	lowestAverageScore := float64(1e9) // Initialize with a very high value

	for distance, totalScore := range distanceScores {
		if distanceCount[distance] > 0 {
			averageScore := totalScore / float64(distanceCount[distance])
			if averageScore < lowestAverageScore {
				lowestAverageScore = averageScore
				preferredDistance = distance
			}
		}
	}

	return preferredDistance
}
