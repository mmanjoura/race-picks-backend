package preparation

import (

	"net/http"
	"strconv"
	"strings"
	"time"

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
	rows, err := db.Query("SELECT selection_link, selection_id, selection_name FROM TodayRunners WHERE  DATE(event_date) = DATE('now')")
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
		// Calculate the average of the BSP and other fields
		selectionsForm, err := getSelectionsForm(selection.Link)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, selectionForm := range selectionsForm {

			result, err := db.ExecContext(c, `
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
				selection.Name, // Include the selection link
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
				time.Now())
			_ = result // Ignore the result if not needed

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

}

func getSelectionsForm(selectionLink string) ([]models.SelectionsForm, error) {
	// Initialize the collector
	c := colly.NewCollector()

	// Slice to store all horse information
	selectionsForm := []models.SelectionsForm{}

	// On HTML element
	c.OnHTML("table.FormTable__StyledTable-sc-1xr7jxa-1 tbody tr", func(e *colly.HTMLElement) {
		raceDate := e.ChildText("td:nth-child(1) a")
		raceLink := e.ChildAttr("td:nth-child(1) a", "href")
		position := e.ChildText("td:nth-child(2)")
		rating := e.ChildText("td:nth-child(3)")
		raceType := e.ChildText("td:nth-child(4)")
		racecourse := e.ChildText("td:nth-child(5)")
		distance := e.ChildText("td:nth-child(6)")
		going := e.ChildText("td:nth-child(7)")
		// class := e.ChildText("td:nth-child(8)")
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
