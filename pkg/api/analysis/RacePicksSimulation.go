package analysis

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
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

	// Example parameters to filter by
	years := strings.Split(raceParams.Years, ",")         //[]string{"2024", "2023", "2022"} // Example year list
	positions := strings.Split(raceParams.Positions, ",") //[]string{"2", "3", "4", "7"} // Example position list
	ages := strings.Split(raceParams.Ages, ",")           //[]string{"2", "3", "4", "5"}      // Example age list

	// Query for today's runners
	rows, err := db.Query(`
		SELECT selection_id,
			   selection_name,
			   event_name,
			   event_date,
			   event_time,
			   price
		FROM EventRunners
		WHERE DATE(event_date) = ? AND event_name = ?`,
		raceParams.EventDate, raceParams.EventName)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var selections []common.Selection
	for rows.Next() {
		var selection common.Selection
		if err := rows.Scan(&selection.ID, &selection.Name, &selection.EventName, &selection.EventDate, &selection.EventTime, &selection.Odds); err != nil {
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
		// Execute the quer

		// Start building the base query
		query := `
					SELECT
						COALESCE(selection_id, 0),
						selection_name,	
						substr(position, 1, 1) as position, 
						Age,
						Trainer,
						Sex,
						Sire,
						Dam,
						Owner,	
						Class,					
						COUNT(*) AS num_runs,
						MAX(race_date) AS last_run_date,
						MAX(race_date) - MIN(race_date) AS duration,
						COUNT(CASE WHEN position = '1' THEN 1 END) AS win_count,
						AVG(CAST(substr(position, 1, 1) AS INTEGER)) AS avg_position,
						AVG(rating) AS avg_rating,
						AVG(distance) AS avg_distance_furlongs,
						AVG(sp_odds) AS sp_odds,
						GROUP_CONCAT(class, ', ') AS all_classes,
						GROUP_CONCAT(race_type, ', ') AS all_race_types,
						GROUP_CONCAT(position, ', ') AS all_positions,
						GROUP_CONCAT(distance, ', ') AS all_distances,
						GROUP_CONCAT(racecourse, ', ') AS all_racecourses,
						GROUP_CONCAT(DATE(race_date), ', ') AS all_race_dates 
					FROM SelectionsForm
					WHERE selection_id = ?`

		// Initialize a slice for query parameters
		params := []interface{}{selection.ID} // Include the selection.ID first

		// Add dynamic year filtering
		if len(years) > 0 {
			yearPlaceholders := make([]string, len(years))
			for i := range years {
				yearPlaceholders[i] = "?"
				params = append(params, years[i]) // Append year to params
			}
			query += " AND strftime('%Y', race_date) IN (" + strings.Join(yearPlaceholders, ", ") + ")"
		}

		// Add dynamic position filtering
		if len(positions) > 0 {
			positionPlaceholders := make([]string, len(positions))
			for i := range positions {
				positionPlaceholders[i] = "position LIKE ?"
				params = append(params, positions[i]+"%") // Append position with wildcard to params
			}
			query += " AND (" + strings.Join(positionPlaceholders, " OR ") + ")"
		}

		// Add dynamic age filtering
		if len(ages) > 0 {
			agePlaceholders := make([]string, len(ages))
			for i := range ages {
				agePlaceholders[i] = "Age LIKE ?"
				params = append(params, ages[i]+"%") // Append age with wildcard to params
			}
			query += " AND (" + strings.Join(agePlaceholders, " OR ") + ")"
		}

		// Add GROUP BY and ORDER BY clauses
		query += `
    GROUP BY selection_id, selection_name, Age, Trainer, Sex, Sire, Dam, Owner, Class
    ORDER BY race_date DESC`

		// Execute the query with dynamically constructed query string and parameters
		rows, err := db.Query(query, params...)
		if err != nil {
			log.Fatal(err)
		}

		if err != nil {
			log.Fatal(err)
		}
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
				&data.AllClasses,
				&data.AllRaceTypes,
				&data.AllPositions,
				&data.AllDistances,
				&data.AllCources,
				&data.AllRaceDates,
			)
			if err != nil {
				continue
			}
		}
		analysisData = append(analysisData, data)
	}

	mapResult := make(map[int64]models.SelectionResult)
	var sortedResults []models.SelectionResult

	result := models.SelectionResult{}

	for id, selecion := range selections {

		if selecion.ID == int64(analysisData[id].SelectionID) {
			totalScore := ScoreSelection(analysisData[id], raceParams, selectionCount[0].NumberOfRuns)
			result.EventName = selecion.EventName
			result.EventTime = selecion.EventTime
			result.SelectionName = selecion.Name
			result.Odds = selecion.Odds
			result.Trainer = analysisData[id].Trainer
			result.AvgPosition = math.Round(analysisData[id].AvgPosition)
			result.AvgRating = math.Round(analysisData[id].AvgRating)
			result.TotalScore = totalScore
			result.Age = analysisData[id].Age
			result.RunCount = analysisData[id].NumRuns
			mapResult[selecion.ID] = result

		}
		if result.TotalScore > 0 {
			sortedResults = append(sortedResults, result)
		}
	}

	// Step 2: Sort the slice by TotalScore
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].TotalScore > sortedResults[j].TotalScore
	})

	// Remove duplicate selection.Name
	var uniqueResults []models.SelectionResult
	uniqueNames := make(map[string]bool)
	for _, result := range sortedResults {
		if !uniqueNames[result.EventTime] {
			uniqueNames[result.EventTime] = true
			uniqueResults = append(uniqueResults, result)
		}
	}

	c.JSON(http.StatusOK, gin.H{"simulationResults": uniqueResults})
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

	for _, race := range races {

		fmt.Print(race)

	}

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

// Helper function to parse race distance considering miles, furlongs, and yards
func parseDistance(dist string) float64 {
	var totalFurlongs float64

	// Split the distance string into components (miles, furlongs, yards)
	parts := strings.Split(dist, " ")

	for _, part := range parts {
		if strings.HasSuffix(part, "m") { // Handle miles
			miles, _ := strconv.ParseFloat(strings.TrimSuffix(part, "m"), 64)
			totalFurlongs += miles * 8.0 // 1 mile = 8 furlongs
		} else if strings.HasSuffix(part, "f") { // Handle furlongs
			furlongs, _ := strconv.ParseFloat(strings.TrimSuffix(part, "f"), 64)
			totalFurlongs += furlongs
		} else if strings.HasSuffix(part, "y") { // Handle yards
			yards, _ := strconv.ParseFloat(strings.TrimSuffix(part, "y"), 64)
			totalFurlongs += yards / 220.0 // 1 furlong = 220 yards
		}
	}

	return totalFurlongs
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

	// 1. Class Analysis
	if strings.Contains(selection.AllClasses, "1") {
		score += 5
	}
	if strings.Contains(selection.AllClasses, "2") {
		score += 3
	}
	if strings.Contains(selection.AllClasses, "3") {
		score += 2
	}

	if strings.Contains(selection.AllRaceTypes, "HURDLE") {
		score += 5
	}
	if strings.Contains(selection.AllRaceTypes, "FLAT") {
		score += 3
	}

	// 2. Cousre Analysis
	strRaceCourses := strings.Split(selection.AllCources, ",")
	var limitedCourses []string
	if len(strRaceCourses) >= limit {
		limitedCourses = strRaceCourses[:limit]
	} else {
		limitedCourses = strRaceCourses[:len(strRaceCourses)]
	}

	for _, course := range limitedCourses {
		course = strings.TrimSpace(course)
		courseScore, err := fetchConstantScore(database.Database.DB, "Course", course)
		if err != nil {
			score -= 1
		}
		score += courseScore

	}

	// 3. Race Distance Analysis
	strDistances := strings.Split(selection.AllDistances, ",")
	var limitedDistances []string
	if len(strDistances) >= limit {
		limitedDistances = strDistances[:limit]
	} else {
		limitedDistances = strDistances[:len(strDistances)]
	}
	var floadDistance float64
	for _, distance := range limitedDistances {
		distance = strings.Trim(distance, " ")
		strDistance := convertDistance(distance)
		fd, _ := strconv.ParseFloat(strDistance, 64)
		floadDistance += fd

	}
	avrDistance := floadDistance / float64(len(strDistances))

	// 4. Race Type Match and Distance Analysis
	if strings.EqualFold(params.RaceType, "HURDLE") {

		distanceDiff := math.Abs(avrDistance - selection.AvgDistanceFurlongs)
		if distanceDiff < 3.2 {
			score += 5 // High score for close distance match
		}
	}
	if strings.EqualFold(params.RaceType, "FLAT") {
		distanceDiff := math.Abs(avrDistance - selection.AvgDistanceFurlongs)
		if distanceDiff < 2.2 {
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
		limitedPostion = strPostion[:len(strPostion)]
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
