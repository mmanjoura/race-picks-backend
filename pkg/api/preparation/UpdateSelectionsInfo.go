package preparation

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

// Mutex for controlling database access
// var dbMutex sync.Mutex

func UpdateSelectionsInfo(c *gin.Context) {
	db := database.Database.DB

	// Get all Selections from eventRunners table
	rows, err := db.Query(`select selection_id, selection_link from EventRunners`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	selections := []models.Selection{}
	for rows.Next() {
		var selection models.Selection
		if err := rows.Scan(&selection.ID, &selection.Link); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		selections = append(selections, selection)
	}

	for _, selection := range selections {

		// Fetch horse information
		horseInformations, err := getForm(selection.Link)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = save(
			db,
			selection.ID,
			horseInformations.Age,
			horseInformations.Trainer,
			horseInformations.Sex,
			horseInformations.Sire,
			horseInformations.Dam,
			horseInformations.Owner,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Horse information saved successfully"})
}

func save(db *sql.DB, selectionID int, age, trainer, sex, sire, dam, owner string) error {

	// Update the SelectionsForm table
	_, err := db.Exec(`
		UPDATE SelectionsForm
		SET Age = ?,
			Trainer = ?,
			Sex = ?,
			Sire = ?,
			Dam = ?,
			Owner = ?
		WHERE selection_id = ?`,
		age, trainer, sex, sire, dam, owner, selectionID,
	)
	if err != nil {
		return err
	}

	return nil
}

func getForm(selectionLink string) (models.SelectionsForm, error) {
	c := colly.NewCollector()

	selectionForm := models.SelectionsForm{}

	c.OnHTML("table.Header__DataTable-xeaizz-1", func(e *colly.HTMLElement) {
		age := e.ChildText("tr:nth-child(1) td.Header__DataValue-xeaizz-4")
		trainer := e.ChildText("tr:nth-child(2) td.Header__DataValue-xeaizz-4 a")
		sex := e.ChildText("tr:nth-child(3) td.Header__DataValue-xeaizz-4")
		sire := e.ChildText("tr:nth-child(4) td.Header__DataValue-xeaizz-4")
		dam := e.ChildText("tr:nth-child(5) td.Header__DataValue-xeaizz-4")
		owner := e.ChildText("tr:nth-child(6) td.Header__DataValue-xeaizz-4")

		// Populate the SelectionsForm object with the scraped data
		selectionForm = models.SelectionsForm{
			Age:     age,
			Trainer: trainer,
			Sex:     sex,
			Sire:    sire,
			Dam:     dam,
			Owner:   owner,
		}
	})

	// Start scraping the URL
	c.Visit("https://www.sportinglife.com" + selectionLink)

	return selectionForm, nil
}
