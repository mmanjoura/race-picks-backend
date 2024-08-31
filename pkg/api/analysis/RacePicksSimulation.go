package analysis

import (
	"database/sql"
	"fmt"
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

	// Query for today's runners
	rows, err := db.Query(`
		SELECT selection_id,
			   selection_name,
			   event_name,
			   event_date,
			   event_time,
			   price
		FROM EventRunners
		WHERE DATE(event_date) = ? AND event_name = ? AND event_time = ?`,
		raceParams.EventDate, raceParams.EventName, raceParams.EventTime)

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
				Class,					
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
				SelectionsForm	WHERE selection_id = ?  order by race_date desc limit ?`, selection.ID, selectionCount[0].NumberOfRuns)
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
		analysisData = append(analysisData, data)
	}

	//  Order this slice by slectionId
	// sort.Slice(selectonsForm, func(i, j int) bool {
	// 	return selectonsForm[i].SelectionID < selectonsForm[j].SelectionID
	// })

	mapResult := make(map[int64]models.SelectionResult)
	var sortedResults []models.SelectionResult

	result := models.SelectionResult{}
	for id, selecion := range selections {

		if selecion.ID == int64(analysisData[id].SelectionID) {
			totalScore := ScoreSelection(analysisData[id], raceParams)
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
	}

	for _, result := range mapResult {
		sortedResults = append(sortedResults, result)
	}

	// Step 2: Sort the slice by TotalScore
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].TotalScore > sortedResults[j].TotalScore
	})

	c.JSON(http.StatusOK, gin.H{"simulationResults": sortedResults})
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
func ScoreSelection(selection models.AnalysisData, params models.RaceParameters) float64 {
	var score float64

	// 1. Race Type Match
	if strings.EqualFold(params.RaceType, "HURDLE") {
		score += 10 // Increase score if the race type matches the preferred type
	}
	if strings.EqualFold(params.RaceType, "FLAT") {
		score += 10 // Increase score if the race type matches the preferred type
	}

	// 2. Race Distance Similarity
	floadDistance, _ := strconv.ParseFloat(params.RaceDistance, 64)
	distanceDiff := math.Abs(floadDistance - selection.AvgDistanceFurlongs)
	if distanceDiff < 2.2 {
		score += 20 // High score for close distance match
	} else if distanceDiff < 3.0 {
		score += 10 // Medium score for a decent distance match
	} else {
		score -= 5 // Penalty for a large distance mismatch
	}

	//  split the selection age and take first element
	intAge, _ := strconv.Atoi(strings.Split(selection.Age, " ")[0])

	if selection.AvgDistanceFurlongs > 12.0 {

		if intAge > 5.0 {
			score += 10 // Increase score if the horse is older than 6 years

		} else {
			score -= 5 // Penalty for class mismatch
		}
	}

	if selection.AvgDistanceFurlongs < 12 {

		if intAge < 5 {
			score += 10 // Increase score if the horse is older than 6 years
		} else {
			score -= 5 // Penalty for class mismatch
		}
	}

	// 6. Recent Performance (Duration since last run)
	if selection.Duration == 4 {
		score += 10 // Good recent performance if the last run was within 30 days
	} else if selection.Duration == 5 {
		score += 5 // Moderate recent performance if the last run was within 60 days
	} else {
		score -= 5 // Penalty for poor recent performance
	}
	// 7. Consistency (Average Position)
	if selection.AvgPosition < 4.0 {
		score += 20 // High score for good average position
	} else if selection.AvgPosition < 6.0 {
		score += 10 // Moderate score for average position
	} else {
		score -= 5 // Penalty for poor average position
	}

	// 8. Winning Potential (Number of Wins)
	if selection.WinCount > 4 {
		score += 15 // High score for many wins
	} else if selection.WinCount > 2 {
		score += 7 // Moderate score for a few wins
	} else {
		score -= 5 // Penalty for low win count
	}

	// 9 Average rating
	if selection.AvgRating > 50 {
		score += 10 // High score for good average rating

	} else {
		score -= 5 // Penalty for low average rating
	}

	// Final score return
	return score
}
