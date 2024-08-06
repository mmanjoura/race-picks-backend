package preparation

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

/// ScrapeRacesInfo godoc
// @Summary Scrape today meeting from the website
// @Description Scrape today meeting from the website
// @Tags ScrapeRacesInfo
// @Accept  json
// @Produce  json
// @Success 200 {object} ScrapeRacesInfo
// @Router /analytics/ScrapeRacesInfo [POST]

func ScrapeRacesInfo(c *gin.Context) {
	db := database.Database.DB

	horseInforamtions, err := getInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, horseInforamtion := range horseInforamtions {
		selectionId, err := getSelectionId(c, horseInforamtion.SelectionName)
		if err != nil {
			if err.Error() == fmt.Sprintf("no horse found with the name: %s", horseInforamtion.SelectionName) {
				continue
			}
		}
		// Do not save the horse information if the selectionId is 0
		// if selectionId == 0 {
		// 	continue
		// }

		// Save horse information to DB
		result, err := db.ExecContext(c, `
		INSERT INTO TodayRunners (
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
			selectionId,
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
	}

	c.JSON(http.StatusOK, gin.H{"message": "Horse information saved successfully"})
}


func getSelectionId(c *gin.Context, horseName string) (int, error) {
	db := database.Database.DB

	var selectionId int
	err := db.QueryRowContext(c, "SELECT selection_id FROM MarketData WHERE selection_name = ?", horseName).Scan(&selectionId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0, nil
		}
		return 0, err
	}

	return selectionId, nil
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

		horse := models.TodayRunners{
			SelectionName: name,
			SelectionLink: selectionLink, // Add the selection link to the struct
			EventTime:     eventTime,
			EventName:     eventName,
			Price:         price,
		}

		horses = append(horses, horse)
	})

	// Start scraping the URL
	c.Visit("https://www.sportinglife.com/racing/abc-guide/tomorrow")
	// c.Visit("https://www.sportinglife.com/racing/abc-guide/abc-guide")

	return horses, nil
}

