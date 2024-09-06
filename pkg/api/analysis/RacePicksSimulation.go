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
func RacePicksSimulation(c *gin.Context) {
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
		if err := rows.Scan(
			&selection.ID, 
			&selection.Name, 
			&selection.EventName, 
			&selection.EventDate, 
			&selection.EventTime, 
			&selection.Odds,
			&selection.RaceDistance,
			&selection.RaceCategory,
			&selection.TrackCondition,
			&selection.NumberOfRunners,
			&selection.RaceTrack,
			&selection.RaceClass,			
			); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		selections = append(selections, selection)
	}

	// Get the selection with the least number of runs
	selectionCount, err := getSelectionWithLeastRuns(db, raceParams.EventName, raceParams.EventTime, raceParams.EventDate)
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

	leastRuns := selectionCount[0].NumberOfRuns
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

			averagePostion := calculateAveragePosition(analysisData[id].AllPositions, leastRuns)
			totalScore := ScoreSelection(analysisData[id], raceParams, leastRuns)
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

	// Get the highest score selection for each event time
	// highestScoresByTime := getHighestScoreByTime(sortedResults)
	// highestScoresByTime := getSecondHighestScoreByTime(sortedResults)
	top3HighestScores := getTop3ScoresByTime(sortedResults)

	// Group the results by EventName
	// groupedResults := make(map[string][]models.SelectionResult)
	// for _, result := range highestScoresByTime {
	// 	groupedResults[result.EventName] = append(groupedResults[result.EventName], result)
	// }

	c.JSON(http.StatusOK, gin.H{"simulationResults": top3HighestScores})
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

	// 3. Race Distance Analysis
	strDistances := strings.Split(selection.AllDistances, ",")
	var limitedDistances []string
	if len(strDistances) >= limit {
		limitedDistances = strDistances[:limit]
	} else {
		limitedDistances = strDistances[:]
	}
	var floadDistance float64
	for _, distance := range limitedDistances {
		distance = strings.Trim(distance, " ")
		strDistance := convertDistance(distance)
		fd, _ := strconv.ParseFloat(strDistance, 64)
		floadDistance += fd

	}
	avrDistance := floadDistance / float64(limit)
	age := ageInString(selection.Age)
	// 4. Race Type Match and Distance Analysis

	if avrDistance <= 8.0 {

		if age == "3" || age == "4" || age == "5" {
			score += 10
		}

		distanceDiff := math.Abs(avrDistance - selection.CurrentDistance)
		if distanceDiff <= 1.0 && distanceDiff >= 0.1 {
			score += 30 // High score for close distance match
		}
		if distanceDiff <= 2.0 && distanceDiff >= 1.1 {
			score += 15 // High score for close distance match
		}
		if distanceDiff <= 2.5 && distanceDiff >= 1.5 {
			score += 10 // High score for close distance match
		}
		if distanceDiff <= 3 && distanceDiff >= 2.6 {
			score += 8 // High score for close distance match
		}
		if distanceDiff <= 4 && distanceDiff >= 3.1 {
			score += 5 // High score for close distance match
		}
	} else {
		if age == "6" || age == "7" || age == "8" {
			score += 10
		}
		distanceDiff := math.Abs(avrDistance - selection.CurrentDistance)
		if distanceDiff <= 3.0 && distanceDiff >= 0.1 {
			score += 30 // High score for close distance match
		}
		if distanceDiff <= 4.0 && distanceDiff >= 3.1 {
			score += 15 // High score for close distance match
		}
		if distanceDiff <= 4.5 && distanceDiff >= 3.5 {
			score += 10 // High score for close distance match
		}
		if distanceDiff <= 5 && distanceDiff >= 4.6 {
			score += 8 // High score for close distance match
		}
		if distanceDiff <= 6 && distanceDiff >= 5.1 {
			score += 5 // High score for close distance match
		}

	}
	// Postion Analysis
	strPostion := strings.Split(selection.AllPositions, ",")

	for _, postion := range strPostion {
		if strings.Contains(postion, "F") {
			score -= 2
		}
		if strings.Contains(postion, "PU") {
			score -= 2
		}
		if strings.Contains(postion, "U") {
			score -= 2
		}
		if strings.Contains(postion, "R") {
			score -= 2
		}
	}

	var limitedPostion []string

	if len(strPostion) >= limit {

		limitedPostion = strPostion[:limit]
	} else {
		limitedPostion = strPostion[:]
	}

	for _, postion := range limitedPostion {
		if strings.Contains(postion, "/") {

			p := strings.Split(postion, "/")
			numerator, err := strconv.Atoi(strings.TrimSpace(p[0]))
			if err != nil {
				score -= 1
			}
			denominator, err := strconv.Atoi(strings.TrimSpace(p[1]))
			if err != nil {
				score -= 1
			}
			score += math.Round(safeDivide(float64(denominator), float64(numerator)))
		}
	}

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