package preparation

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

func SaveSelectionsForm(c *gin.Context) {
	db := database.Database.DB

	type Selection struct {
		ID   int
		Name string
		Link string
	}
	
	var params models.TodayRunners

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get today's runners for the given event_name and event_date
	rows, err := db.Query("SELECT selection_link, selection_id, selection_name FROM TodayRunners WHERE DATE(event_date) = DATE('now')")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Create a slice to hold the selections
	var selections []Selection
	for rows.Next() {
		var selection Selection
		err := rows.Scan(&selection.Link, &selection.ID, &selection.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		selections = append(selections, selection)
	}

	for _, selection := range selections {
		// Scrape and clean the data
		selectionsForm, err := getSelectionsForm(selection.Link)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Process each entry in selectionsForm
		for _, selectionForm := range selectionsForm {

			// Clean and transform the data
			selectionForm.Position = strconv.Itoa(extractPosition(selectionForm.Position))
			selectionForm.Distance = fmt.Sprintf("%.1f", convertDistance(selectionForm.Distance))
			selectionForm.Going = standardizeGoing(selectionForm.Going)
			selectionForm.SPOdds = fmt.Sprintf("%.2f", oddsToProbability(selectionForm.SPOdds))

			// Insert cleaned data into the database
			_, err := db.ExecContext(c, `
				INSERT INTO SelectionsForm (
					selection_name,
					selection_id,
					race_date,
					position,
					rating,
					race_type,
					racecourse,
					distance,
					going,
					draw,
					sp_odds,
					created_at
				)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				selection.Name,
				selection.ID,
				selectionForm.RaceDate,
				selectionForm.Position,
				selectionForm.Rating,
				selectionForm.RaceType,
				selectionForm.Racecourse,
				selectionForm.Distance,
				selectionForm.Going,
				selectionForm.Draw,
				selectionForm.SPOdds,
				time.Now(),
			)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}
}

func getSelectionsForm(selectionLink string) ([]models.SelectionsForm, error) {
	c := colly.NewCollector()

	// Slice to store all horse information
	selectionsForm := []models.SelectionsForm{}

	c.OnHTML("table.FormTable__StyledTable-sc-1xr7jxa-1 tbody tr", func(e *colly.HTMLElement) {
		raceDate := e.ChildText("td:nth-child(1) a")
		raceLink := e.ChildAttr("td:nth-child(1) a", "href")
		position := e.ChildText("td:nth-child(2)")
		rating := e.ChildText("td:nth-child(3)")
		raceType := e.ChildText("td:nth-child(4)")
		racecourse := e.ChildText("td:nth-child(5)")
		distance := e.ChildText("td:nth-child(6)")
		going := e.ChildText("td:nth-child(7)")
		spOdds := e.ChildText("td:nth-child(9)")

		// Parsing the rating into an integer
		ratingValue := 0
		if rating != "" {
			ratingValue = parseInt(rating)
		}

		// Parsing the draw (assuming it's contained in position if applicable)
		drawValue := 0
		if position != "" {
			drawValue = parseInt(position)
		}

		// Parsing race date to time.Time
		parsedDate, _ := time.Parse("02/01/06", raceDate) // Assuming UK date format

		// Split the date by "/" and add the current year
		dateParts := strings.Split(raceDate, "/")
		raceDate = "20" + dateParts[2] + "-" + dateParts[1] + "-" + dateParts[1]  

		selectionForm := models.SelectionsForm{
			RaceDate:   raceDate,
			Position:   position,
			Rating:     ratingValue,
			RaceType:   raceType,
			Racecourse: racecourse,
			Distance:   distance,
			Going:      going,
			Draw:       drawValue,
			SPOdds:     spOdds,
			RaceURL:    raceLink,
			EventDate:  parsedDate,
			CreatedAt:  time.Now(),
		}

		selectionsForm = append(selectionsForm, selectionForm)
	})

	// Start scraping the URL
	c.Visit("https://www.sportinglife.com" + selectionLink)

	return selectionsForm, nil
}

// Helper function to parse a string to an int
func parseInt(value string) int {
	parsedValue, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsedValue
}

// extractPosition extracts the numeric part of the position (e.g., "10/12" -> 10).
func extractPosition(posStr string) int {
	parts := strings.Split(posStr, "/")
	if len(parts) > 0 {
		pos, err := strconv.Atoi(parts[0])
		if err == nil {
			return pos
		}
	}
	return 0
}

// convertDistance converts a distance string like "2m 1f 47y" to furlongs.
func convertDistance(distanceStr string) float64 {
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
	return furlongs
}

// oddsToProbability converts fractional odds (e.g., "5/2") to a probability.
func oddsToProbability(oddsStr string) float64 {
	parts := strings.Split(oddsStr, "/")
	if len(parts) == 2 {
		numerator, err1 := strconv.ParseFloat(parts[0], 64)
		denominator, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 == nil && err2 == nil {
			return numerator / (numerator + denominator)
		}
	}
	return 0.0
}

// standardizeGoing converts various going terms into a standard format
func standardizeGoing(goingStr string) string {
	switch goingStr {
	case "Good to Firm", "Good (Good to Firm in places)":
		return "GoodToFirm"
	case "Standard / Slow":
		return "StandardSlow"
	default:
		return goingStr
	}
}
