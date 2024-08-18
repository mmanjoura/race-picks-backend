package preparation

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

type Selection struct {
	ID   int
	Name string
	Link string
}

func ScrapeRacesInfo(c *gin.Context) {
	db := database.Database.DB

	horseInforamtions, err := getInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, horseInforamtion := range horseInforamtions {

		// Save horse information to DB
		result, err := db.ExecContext(c, `
		INSERT INTO EventRunners (
			selection_link,
			selection_id,
			selection_name,	
			event_time,
			event_name,
			price,		
			event_date,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			horseInforamtion.SelectionLink, // Include the selection link
			horseInforamtion.SelectionID,
			horseInforamtion.SelectionName,
			horseInforamtion.EventTime,
			horseInforamtion.EventName,
			horseInforamtion.Price,
			time.Now(),
			time.Now())
		_ = result // Ignore the result if not needed
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = saveSelectionsForm(c, horseInforamtion.SelectionID, horseInforamtion.SelectionLink, horseInforamtion.SelectionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	}

	c.JSON(http.StatusOK, gin.H{"message": "Horse information saved successfully"})
}

func getInfo() ([]models.TodayRunners, error) {
	// Initialize the collector
	c := colly.NewCollector()

	// Slice to store all horse information
	horses := []models.TodayRunners{}

	// On HTML element
	c.OnHTML("table.AbcTable__TableContent-sc-9z9a8v-3 tbody tr", func(e *colly.HTMLElement) {
		name := e.ChildText("td:nth-child(1) a")
		selectionLink := e.ChildAttr("td:nth-child(1) a", "href") // Get the href attribute
		event := e.ChildText("td:nth-child(3) a")
		price := e.ChildText("th:nth-child(5) span")

		parts := strings.SplitN(event, " ", 2)
		eventTime := parts[0]
		eventName := parts[1]

		// Compile the regular expression to match digits at the end of the string
		re := regexp.MustCompile(`/horse/(\d+)$`)

		// Find the substring that matches the pattern and extract the horse_Id
		match := re.FindStringSubmatch(selectionLink)
		selectionId := 0
		if len(match) > 1 {
			selectionId, _ = strconv.Atoi(match[1])

		}
		horse := models.TodayRunners{
			SelectionName: name,
			SelectionLink: selectionLink, // Add the selection link to the struct
			EventTime:     eventTime,
			EventName:     eventName,
			Price:         cleanString(price),
			SelectionID:   selectionId,
		}

		horses = append(horses, horse)
	})

	// Start scraping the URL
	// c.Visit("https://www.sportinglife.com/racing/abc-guide/tomorrow")
	c.Visit("https://www.sportinglife.com/racing/abc-guide/abc-guide")

	return horses, nil
}

func cleanString(input string) string {
	// Split the string by "/"
	parts := strings.Split(input, "/")

	// Check if there are at least two parts
	if len(parts) >= 2 {
		// Keep only the first and last part
		parts = []string{parts[0], parts[len(parts)-1]}
	}

	// Join the parts back together with "/"
	return strings.Join(parts, "/")
}

func saveSelectionsForm(c *gin.Context, selectionID int, selectionLink, selectionName string) error {
	db := database.Database.DB

	// first check if this selection exist in SelectionsForm table
	// if it does, then get on the missing seletion Form data
	rows, err := db.Query(`
			select selection_id, race_date  from SelectionsForm where selection_id = ? order by race_date desc limit 1`, selectionID)

	if err != nil {
		return err
	}
	// scan the rows
	var raceDate time.Time
	var selection_id int
	for rows.Next() {
		err := rows.Scan(&selection_id, &raceDate)
		if err != nil {
			return err
		}
	}

	defer rows.Close()

	if selection_id == 0 {
		// Scrape and clean the data
		selectionsForm, err := getAllSelectionsForm(selectionLink)
		if err != nil {
			return err
		}
		err = saveSelectionForm(selectionsForm, c, selectionName, selectionID)
		if err != nil {
			return err
		}

	} else {
		// Get the last date of the selection form
		selectionsForm, err := getLatestSelectionsForm(selectionLink, raceDate)

		if err != nil {
			return err
		}

		err = saveSelectionForm(selectionsForm, c, selectionName, selectionID)

		if err != nil {
			return err

		}
	}
	return nil
}

func saveSelectionForm(selectionsForm []models.SelectionsForm, c *gin.Context, selectionName string, selectionID int) error {
	db := database.Database.DB
	for _, selectionForm := range selectionsForm {

		selectionForm.Position = strconv.Itoa(extractPosition(selectionForm.Position))
		selectionForm.Distance = fmt.Sprintf("%.1f", convertDistance(selectionForm.Distance))
		selectionForm.Going = standardizeGoing(selectionForm.Going)
		selectionForm.SPOdds = fmt.Sprintf("%.2f", odds(selectionForm.SPOdds))

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
					class,
					sp_odds,
					created_at
				)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			selectionName,
			selectionID,
			selectionForm.RaceDate,
			selectionForm.Position,
			selectionForm.Rating,
			selectionForm.RaceType,
			selectionForm.Racecourse,
			selectionForm.Distance,
			selectionForm.Going,
			selectionForm.Class,
			selectionForm.SPOdds,
			time.Now(),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}
	}
	return nil
}

func getAllSelectionsForm(selectionLink string) ([]models.SelectionsForm, error) {
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
		class := e.ChildText("td:nth-child(8)")
		spOdds := e.ChildText("td:nth-child(9)")

		// Parsing the rating into an integer
		ratingValue := 0
		if rating != "" {
			ratingValue = parseInt(rating)
		}

		classValue := 0
		if class != "" {
			classValue, _ = strconv.Atoi(class)
		}

		// Parsing race date to time.Time
		parsedDate, _ := time.Parse("02/01/06", raceDate) // Assuming UK date format

		// Split the date by "/" and add the current year
		dateParts := strings.Split(raceDate, "/")
		raceDate = "20" + dateParts[2] + "-" + dateParts[1] + "-" + dateParts[0]
		// convert raceDate to time.Time
		parsedRaceDate, _ := time.Parse("2006-01-02", raceDate)

		selectionForm := models.SelectionsForm{
			RaceDate:   parsedRaceDate,
			Position:   position,
			Rating:     ratingValue,
			RaceType:   raceType,
			Racecourse: racecourse,
			Distance:   distance,
			Going:      going,
			Class:       classValue,
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

func getLatestSelectionsForm(selectionLink string, lasRuntDate time.Time) ([]models.SelectionsForm, error) {
	c := colly.NewCollector()

	// Slice to store all horse information
	selectionsForm := []models.SelectionsForm{}

	// Flag to indicate if the first row has been processed
	// firstRowProcessed := false

	c.OnHTML("table.FormTable__StyledTable-sc-1xr7jxa-1 tbody tr", func(e *colly.HTMLElement) {
		// if firstRowProcessed {
		// 	return
		// }

		raceDate := e.ChildText("td:nth-child(1) a")
		raceLink := e.ChildAttr("td:nth-child(1) a", "href")
		position := e.ChildText("td:nth-child(2)")
		rating := e.ChildText("td:nth-child(3)")
		raceType := e.ChildText("td:nth-child(4)")
		racecourse := e.ChildText("td:nth-child(5)")
		distance := e.ChildText("td:nth-child(6)")
		going := e.ChildText("td:nth-child(7)")
		class := e.ChildText("td:nth-child(8)")
		spOdds := e.ChildText("td:nth-child(9)")

		// Parsing the rating into an integer
		ratingValue := 0
		if rating != "" {
			ratingValue = parseInt(rating)
		}

		classValue := 0
		if class != "" {
			classValue, _ = strconv.Atoi(class)
		}
	

		// Split the date by "/" and add the current year
		dateParts := strings.Split(raceDate, "/")
		raceDate = "20" + dateParts[2] + "-" + dateParts[1] + "-" + dateParts[0]

	

		// Convert raceDate to time.Time
		parsedRaceDate, _ := time.Parse("2006-01-02", raceDate)
		fmt.Println(" Screaped date: %s", parsedRaceDate)
		fmt.Println(" Existing Date: %s", lasRuntDate)

		if parsedRaceDate.Before(lasRuntDate) {
			return
		}

		// Create a new SelectionsForm object with the scraped data
		selectionForm := models.SelectionsForm{
			RaceDate:   parsedRaceDate,
			Position:   position,
			Rating:     ratingValue,
			RaceType:   raceType,
			Racecourse: racecourse,
			Distance:   distance,
			Going:      going,
			Class:      classValue,
			SPOdds:     spOdds,
			RaceURL:    raceLink,
			EventDate:  parsedRaceDate,
			CreatedAt:  time.Now(),
		}

		// Append the selection form to the slice
		selectionsForm = append(selectionsForm, selectionForm)

		// Set the flag indicating that the first row has been processed
		// firstRowProcessed = true

		// Abort the request after processing the first row
		e.Request.Abort()
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

// odds converts fractional odds (e.g., "5/2") to a probability.
func odds(oddsStr string) float64 {
	parts := strings.Split(oddsStr, "/")
	if len(parts) == 2 {
		numerator, err1 := strconv.ParseFloat(parts[0], 64)
		denominator, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 == nil && err2 == nil {
			return numerator / (denominator)
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