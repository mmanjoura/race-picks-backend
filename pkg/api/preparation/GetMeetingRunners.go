package preparation

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

type DaySince struct {
	RaceDate    time.Time `json:"race_date"`
	SelectionID int       `json:"selection_id"`
}

type Diff struct {
	DaySince
	DateDiffInDays int `json:"date_diff_in_days"`
}

func GetMeetingRunners(c *gin.Context) {
	db := database.Database.DB
	type Selection struct {
		ID        int
		Name      string
		EventDate time.Time
	}

	eventName := c.Query("event_name")
	eventTime := c.Query("event_time")
	eventDate := c.Query("event_date")


	// Get distinct Event name from SelectionsForm table
	rows, err := db.Query(`
		SELECT DISTINCT racecourse FROM SelectionsForm
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var eventNames []string	
	for rows.Next() {
		var eventName string
		err := rows.Scan(
			&eventName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		eventNames = append(eventNames, eventName)
	}


	
	// Get today's runners for the given event_name and event_date
	rows, err = db.Query(`
			SELECT 	selection_id, 
					selection_name, 
					event_date FROM 
					EventRunners WHERE  
					event_name = ? and event_time = ? and DATE(event_date) = ?`,
		eventName, eventTime, eventDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var analysisData []models.AnalysisData
	var selections []Selection
	for rows.Next() {
		var selection Selection
		err := rows.Scan(
			&selection.ID,
			&selection.Name,
			&selection.EventDate,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if selection.ID == 0 {
			continue
		}
		selections = append(selections, selection)
	}

	for _, selection := range selections {

		// Execute the query
		rows, err = db.Query(`
						SELECT
							selection_id,
							selection_name,							
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
							SelectionsForm	WHERE selection_id = ? order by race_date desc`, selection.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		// Loop through the rows and append the results to the slice
		var data models.AnalysisData

		for rows.Next() {
			err := rows.Scan(
				&data.SelectionID,
				&data.SelectionName,
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

	for i, data := range analysisData {
		recoveryDays, err := getRecoveryDays(data.SelectionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		analysisData[i].RecoveryDays = recoveryDays

		// Get Analysis trend

		dates := strings.Split(data.AllRaceDates, ",")
		distances := strings.Split(data.AllDistances, ",")
		positions := strings.Split(data.AllPositions, ",")
		events := strings.Split(data.AllCources, ",")
		selectionForm, err := parseData(dates, distances, positions, events)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		analysis := analyzeTrends(selectionForm)
		analysisData[i].TrendAnalysis = analysis


	}

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the meeting data
	c.JSON(http.StatusOK, gin.H{"meetingData": analysisData})
}

func getRecoveryDays(selectionID int) (int, error) {
	db := database.Database.DB

	// Execute the query with WITH daysSince as subquery
	rows, err := db.Query(`
		select 	race_date, 
				selection_id from   
		SelectionsForm where selection_id = ? 
		order by race_date DESC limit 2;
	`, selectionID)

	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var daysSince []DaySince

	// Iterate over the rows and scan the result into the DaySince struct
	for rows.Next() {
		var race DaySince
		err := rows.Scan(
			&race.RaceDate,
			&race.SelectionID,
		)
		if err != nil {
			return 0, err
		}
		daysSince = append(daysSince, race)
	}

	// Check for errors after iterating through the rows
	if err = rows.Err(); err != nil {
		return 0, err
	}

	// Calculate the difference between the two dates, if there are at least 2 races
	if len(daysSince) == 2 {
		// Normalize dates by removing the time component
		raceDate1 := time.Date(daysSince[0].RaceDate.Year(), daysSince[0].RaceDate.Month(), daysSince[0].RaceDate.Day(), 0, 0, 0, 0, time.UTC)
		raceDate2 := time.Date(daysSince[1].RaceDate.Year(), daysSince[1].RaceDate.Month(), daysSince[1].RaceDate.Day(), 0, 0, 0, 0, time.UTC)

		dateDiff := int(raceDate1.Sub(raceDate2).Hours() / 24)

		return dateDiff, nil
	}

	// If there is only one race or none, we cannot calculate a meaningful difference
	return 0, nil
}

// analyzeTrends analyzes the race data and returns an AnalyzeTrends struct with the results
func analyzeTrends(raceData []models.RaceData) models.AnalyzeTrends {
	var bestDistances []float64
	var bestRaces []models.RaceData

	for _, race := range raceData {
		if race.Position <= 3 {
			bestDistances = append(bestDistances, race.Distance)
			bestRaces = append(bestRaces, race)
		}
	}

	if len(bestDistances) == 0 {
		return models.AnalyzeTrends{}
	}

	// Determine the optimal distance range
	minDistance := bestDistances[len(bestDistances)-1]
	maxDistance := bestDistances[0]

	return models.AnalyzeTrends{
		BestRaces:         bestRaces,
		OptimalDistanceMin: minDistance,
		OptimalDistanceMax: maxDistance,
	}
}

func parseData(dates, distances, positions, events []string) ([]models.RaceData, error) {
	var raceData []models.RaceData

	for i := range dates {
		date, err := time.Parse("2006-01-02", strings.TrimSpace(dates[i]))
		if err != nil {
			return nil, err
		}

		var distance float64
		fmt.Sscanf(distances[i], "%f", &distance)

		var position int
		fmt.Sscanf(positions[i], "%d", &position)

		raceData = append(raceData, models.RaceData{
			Date:     date,
			Distance: distance,
			Position: position,
			Event:    events[i],
		})
	}

	return raceData, nil
}

