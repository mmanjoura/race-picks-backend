package preparation

import (
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

func GetRacingMarketData(c *gin.Context) {
	db := database.Database.DB

	var raceDate models.EventDate

	// Bind JSON input to optimalParams
	if err := c.ShouldBindJSON(&raceDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todayRunners, err := getTodayRunners()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, todayRunner := range todayRunners {

		// Save horse information to DB
		result, err := db.ExecContext(c, `
		INSERT INTO EventRunners (
			selection_link,
			selection_id,
			event_link,
			selection_name,	
			event_time,
			event_name,
			price,		
			event_date,
			race_distance,
			race_category,
			track_condition,
			number_of_runners,
			race_track,
			race_class,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			todayRunner.SelectionLink, // Include the selection link
			todayRunner.SelectionID,
			todayRunner.EventLink,
			todayRunner.SelectionName,
			todayRunner.EventTime,
			todayRunner.EventName,
			todayRunner.Price,
			time.Now(),
			todayRunner.RaceConditon.RaceDistance,
			todayRunner.RaceConditon.RaceCategory,
			todayRunner.RaceConditon.TrackCondition,
			todayRunner.RaceConditon.NumberOfRunners,
			todayRunner.RaceConditon.RaceTrack,
			todayRunner.RaceConditon.RaceClass,
			time.Now())
		_ = result // Ignore the result if not needed
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}


	c.JSON(http.StatusOK, gin.H{"message": "Horse information saved successfully"})

}

func getTodayRunners() ([]models.TodayRunners, error) {
	// Initialize the collector
	c := colly.NewCollector()

	// Slice to store all horse information
	horses := []models.TodayRunners{}

	// On HTML element
	c.OnHTML("table.AbcTable__TableContent-sc-9z9a8v-3 tbody tr", func(e *colly.HTMLElement) {
		name := e.ChildText("td:nth-child(1) a")
		selectionLink := e.ChildAttr("td:nth-child(1) a", "href") // Get the href attribute
		event := e.ChildText("td:nth-child(3) a")
		eventLink := e.ChildAttr("td:nth-child(3) a", "href") // Get the href attribute
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
			EventLink:     eventLink,     // Add the event link to the struct
			EventTime:     eventTime,
			EventName:     eventName,
			Price:         removeDuplicateOdds(price),
			SelectionID:   selectionId,
		}

		reaceConditons := getEventConditons(eventLink)
		horse.RaceConditon = reaceConditons

		horses = append(horses, horse)
	})

	// Start scraping the URL
	c.Visit("https://www.sportinglife.com/racing/abc-guide/abc-guide")

	return horses, nil
}

func getEventConditons(eventLink string) models.RaceConditon {
	// Initialize the collector
	c := colly.NewCollector()
	raceConditons := models.RaceConditon{}

	// Define variables to hold extracted values
	// var raceCategory, raceDistance, trackCondition, numberOfRunners, raceClass, raceTrack string

	// Set the HTML selector and processing logic
	c.OnHTML("li.RacingRacecardSummary__StyledAdditionalInfo-sc-1intsbr-2", func(e *colly.HTMLElement) {
		// Extract the text from the HTML element
		content := e.Text

		raceConditons = extractRaceInfo(content)

	})

	// Visit the URL
	c.Visit("https://www.sportinglife.com" + eventLink)

	return raceConditons
}

func extractRaceInfo(content string) models.RaceConditon {
	// Split the content by '|'
	parts := strings.Split(content, "|")

	// Trim whitespace from each part
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	var (
		raceDistance    string
		numberOfRunners string
	)

	// Patterns to recognize specific parts
	distancePattern := regexp.MustCompile(`\d+m \d+f \d+y|\d+f \d+y|\d+m|\d+f|\d+y`) // Extended pattern for various distance formats
	runnersPattern := regexp.MustCompile(`\d+ Runners`)

	// Assign defaults
	raceDistance = "Unknown"
	numberOfRunners = "Unknown"

	// Assign values based on regex patterns
	for _, part := range parts {
		switch {
		case distancePattern.MatchString(part):
			raceDistance = part
		case runnersPattern.MatchString(part):
			numberOfRunners = part
		}
	}

	// Create the RaceCondition object
	raceConditions := models.RaceConditon{
		RaceDistance:    raceDistance,
		NumberOfRunners: numberOfRunners,
	}

	return raceConditions
}

// Function to remove duplicate odds patterns
func removeDuplicateOdds(odds string) string {
	// Count the number of '/' characters in the string
	count := strings.Count(odds, "/")

	// If there are exactly two '/' characters, return the first half of the string
	if count == 2 {
		mid := len(odds) / 2
		return odds[:mid]
	}

	// If there is not exactly two '/', return the original string
	return odds
}
