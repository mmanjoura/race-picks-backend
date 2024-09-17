package analysis

// import (
// 	"database/sql"
// 	"fmt"
// 	"math"
// 	"net/http"
// 	"sort"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/mmanjoura/race-picks-backend/pkg/database"
// 	"github.com/mmanjoura/race-picks-backend/pkg/models"

// 	"github.com/gin-gonic/gin"
// )

// type DaySince struct {
// 	RaceDate    time.Time `json:"race_date"`
// 	SelectionID int       `json:"selection_id"`
// }

// type Diff struct {
// 	DaySince
// 	DateDiffInDays int `json:"date_diff_in_days"`
// }

// type Selection struct {
// 	ID        int
// 	Name      string
// 	EventDate time.Time
// 	NumberOfRuns int
// }


// func GetMeetingPrediction_(c *gin.Context) {
// 	db := database.Database.DB

// 	analysisDataResponse := models.AnalysisDataResponse{}

// 	var raceParams models.RaceParameters

// 	// Bind JSON input to optimalParams
// 	if err := c.ShouldBindJSON(&raceParams); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	eventName := raceParams.EventName
// 	eventTime := raceParams.EventTime
// 	eventDate := raceParams.EventDate
// 	raceType := raceParams.RaceType




// 	// Get the selection with the least number of runs
// 	selectionCount, err := getSelectionWithLeastRuns(db, eventName, eventTime, eventDate)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	_ = selectionCount

// 	// Get today's runners for the given event_name and event_date
// 	rows, err := db.Query(`
// 			SELECT 	selection_id, 
// 					selection_name, 
// 					event_date FROM 
// 					EventRunners WHERE  
// 					event_name = ? and event_time = ? and DATE(event_date) = ?`,
// 		eventName, eventTime, eventDate)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	var analysisData []models.AnalysisData
// 	var selections []Selection
// 	for rows.Next() {
// 		var selection Selection
// 		err := rows.Scan(
// 			&selection.ID,
// 			&selection.Name,
// 			&selection.EventDate,
// 		)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		if selection.ID == 0 {
// 			continue
// 		}
// 		selections = append(selections, selection)
// 	}

// 	for _, selection := range selections {

// 		// Execute the query
// 		rows, err = db.Query(`
// 			SELECT
// 				COALESCE(selection_id, 0),
// 				selection_name,	
// 				substr(position, 1, 1) as positon, 
// 				Age,
// 				Trainer,
// 				Sex,
// 				Sire,
// 				Dam,
// 				Owner,	
// 				Class,					
// 				COUNT(*) AS num_runs,
// 				MAX(race_date) AS last_run_date,
// 				MAX(race_date) - MIN(race_date) AS duration,
// 				COUNT(CASE WHEN position = '1' THEN 1 END) AS win_count,
// 				AVG(position) AS avg_position,
// 				AVG(rating) AS avg_rating,
// 				AVG(distance) AS avg_distance_furlongs,
// 				AVG(sp_odds) AS sp_odds,
// 				GROUP_CONCAT(position, ', ') AS all_positions,
// 				GROUP_CONCAT(distance, ', ') AS all_distances,
// 				GROUP_CONCAT(racecourse, ', ') AS all_racecources,
// 				GROUP_CONCAT(DATE(race_date), ', ') AS all_race_dates 
// 			FROM
// 				SelectionsForm	WHERE selection_id = ?  order by race_date desc`, selection.ID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		defer rows.Close()

// 		// Loop through the rows and append the results to the slice
// 		var data models.AnalysisData
// 		noData := false

// 		for rows.Next() {
// 			noData = false
// 			err := rows.Scan(
// 				&data.SelectionID,
// 				&data.SelectionName,
// 				&data.Position,
// 				&data.Age,
// 				&data.Trainer,
// 				&data.Sex,
// 				&data.Sire,
// 				&data.Dam,
// 				&data.Owner,
// 				&data.EventClass,
// 				&data.NumRuns,
// 				&data.LastRunDate,
// 				&data.Duration,
// 				&data.WinCount,
// 				&data.AvgPosition,
// 				&data.AvgRating,
// 				&data.AvgDistanceFurlongs,
// 				&data.AvgOdds,
// 				&data.AllPositions,
// 				&data.AllDistances,
// 				&data.AllCources,
// 				&data.AllRaceDates,
// 			)
// 			if err != nil {
// 				noData = true
// 				continue
// 			}
// 		}
// 		if !noData {

// 			winLose, err := getRaceResult(rows, err, db, eventDate, c, selection.ID)
// 			if err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 				return
// 			}

// 			data.WinLose = winLose

// 			analysisData = append(analysisData, data)

// 		}

// 	} // Sorting logic
// 	sort.Slice(analysisData, func(i, j int) bool {
// 		// Sort by winner positions (1, 2, 3) first
// 		positions := map[int]bool{1: true, 2: true, 3: true}
// 		posI, posJ := analysisData[i].WinLose.Position, analysisData[j].WinLose.Position

// 		pi, err := strconv.Atoi(posI)
// 		if err != nil {
// 			pi = 0
// 		}
// 		pj, err := strconv.Atoi(posJ)
// 		if err != nil {
// 			pj = 0
// 		}

// 		if positions[pi] && !positions[pj] {
// 			return true
// 		} else if !positions[pi] && positions[pj] {
// 			return false
// 		} else if posI != posJ {
// 			return posI < posJ
// 		}

// 		analysisDataResponse.Selections = analysisData

// 		// Then by average position
// 		return analysisData[i].AvgPosition < analysisData[j].AvgPosition
// 	})
// 	results := []models.SelectionResult{}

// 	for i, data := range analysisData {

// 		if i > 0 {
// 			break
// 		}

// 		result := models.SelectionResult{}
// 		result.EventName = eventName
// 		result.EventDate = eventDate
// 		result.EventTime = eventTime
// 		result.SelectionName = data.SelectionName
// 		result.RaceType = raceType
// 		result.EventClass = data.EventClass
// 		result.Odds = strconv.FormatFloat(math.Round(data.AvgOdds), 'f', -1, 64)
// 		result.Trainer = data.Trainer

// 		results = append(results, result)

// 		// First check if we have already insertions for the same event_name, event_date, event_time, selection_name
// 		rows, err = db.Query(`SELECT * FROM EventPredictions WHERE event_name = ? and event_date = ? and event_time = ?`,
// 			result.EventName, result.EventDate, result.EventTime)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		defer rows.Close()

// 		// If we have already inserted the data, we skip the insertion
// 		if rows.Next() {
// 			continue
// 		}
// 		_, err = db.Exec(`
// 				INSERT INTO EventPredictions (event_name, event_date, event_time, selection_name, event_class, race_type, odds, trainer)
// 				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
// 			result.EventName, result.EventDate, result.EventTime, result.SelectionName, result.EventClass, result.RaceType, result.Odds, result.Trainer)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

// 		if i > 0 {

// 			break
// 		}

// 	}

// 	// Return the meeting data
// 	c.JSON(http.StatusOK, gin.H{"analysisDataResponse": results})
// }

// func getRaceResult(rows *sql.Rows, err error, db *sql.DB, eventDate string, c *gin.Context, selectionID int) (models.WinLose, error) {
	
// 	rows, err = db.Query(`
// 		SELECT 	selection_id,
// 				selection_name,
// 				race_date,
// 				SUBSTR(position, 1, INSTR(position, '/') - 1) as postion
// 		FROM SelectionsForm
// 		WHERE DATE(race_date) = ? and selection_id = ?`, eventDate, selectionID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return models.WinLose{}, err
// 	}
// 	defer rows.Close()
// 	var data models.WinLose
// 	for rows.Next() {
// 		err := rows.Scan(
// 			&data.SelectionID,
// 			&data.SelectionName,
// 			&data.EventDate,
// 			&data.Position,
// 		)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return models.WinLose{}, err
// 		}

// 	}
// 	return data, nil
// }

// func getRecoveryDays(selectionID int, eventDate string) (float64, error) {
// 	db := database.Database.DB

// 	// Execute the query with WITH daysSince as subquery
// 	rows, err := db.Query(`
// 		select 	race_date, 
// 				selection_id from   
// 		SelectionsForm where selection_id = ? 
// 		order by race_date DESC limit 2;
// 	`, selectionID)

// 	if err != nil {
// 		return 0, err
// 	}
// 	defer rows.Close()

// 	var daysSince []DaySince

// 	// Iterate over the rows and scan the result into the DaySince struct
// 	for rows.Next() {
// 		var race DaySince
// 		err := rows.Scan(
// 			&race.RaceDate,
// 			&race.SelectionID,
// 		)
// 		if err != nil {
// 			return 0, err
// 		}
// 		daysSince = append(daysSince, race)
// 	}

// 	// Check for errors after iterating through the rows
// 	if err = rows.Err(); err != nil {
// 		return 0, err
// 	}

// 	// Calculate the difference between the two dates, if there are at least 2 races
// 	if len(daysSince) == 2 {
// 		// Normalize dates by removing the time component
// 		var lastRunDate time.Time
// 		if eventDate == time.Now().Format("2006-01-02") {
// 			lastRunDate = time.Date(daysSince[0].RaceDate.Year(), daysSince[0].RaceDate.Month(), daysSince[0].RaceDate.Day(), 0, 0, 0, 0, time.UTC)
// 		} else {
// 			lastRunDate = time.Date(daysSince[1].RaceDate.Year(), daysSince[1].RaceDate.Month(), daysSince[1].RaceDate.Day(), 0, 0, 0, 0, time.UTC)
// 		}

// 		date, err := time.Parse("2006-01-02", eventDate)
// 		if err != nil {
// 			return 0.0, err
// 		}
// 		dateDiff := date.Sub(lastRunDate).Hours() / 24

// 		return math.Abs(dateDiff), nil
// 	}

// 	// If there is only one race or none, we cannot calculate a meaningful difference
// 	return 0, nil
// }

// // analyzeTrends analyzes the race data and returns an AnalyzeTrends struct with the results
// func analyzeTrends(raceData []models.RaceData) models.AnalyzeTrends {
// 	var bestDistances []float64
// 	var bestRaces []models.RaceData

// 	for _, race := range raceData {
// 		if race.Position <= 3 {
// 			bestDistances = append(bestDistances, race.Distance)
// 			bestRaces = append(bestRaces, race)
// 		}
// 	}

// 	if len(bestDistances) == 0 {
// 		return models.AnalyzeTrends{}
// 	}

// 	// Determine the optimal distance range
// 	minDistance := bestDistances[len(bestDistances)-1]
// 	maxDistance := bestDistances[0]

// 	return models.AnalyzeTrends{
// 		BestRaces:          bestRaces,
// 		OptimalDistanceMin: minDistance,
// 		OptimalDistanceMax: maxDistance,
// 	}
// }

// func parseData(dates, distances, positions, events []string) ([]models.RaceData, error) {
// 	var raceData []models.RaceData

// 	for i := range dates {
// 		date, err := time.Parse("2006-01-02", strings.TrimSpace(dates[i]))
// 		if err != nil {
// 			return nil, err
// 		}

// 		var distance float64
// 		fmt.Sscanf(distances[i], "%f", &distance)

// 		var position int
// 		fmt.Sscanf(positions[i], "%d", &position)

// 		raceData = append(raceData, models.RaceData{
// 			Date:     date,
// 			Distance: distance,
// 			Position: position,
// 			Event:    events[i],
// 		})
// 	}

// 	return raceData, nil
// }

// // Calculate average of a slice of floats
// func average(values []float64) float64 {
// 	if len(values) == 0 {
// 		return 0
// 	}
// 	var sum float64
// 	for _, value := range values {
// 		sum += value
// 	}
// 	return sum / float64(len(values))
// }

