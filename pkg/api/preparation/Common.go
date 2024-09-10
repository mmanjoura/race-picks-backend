package preparation

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

func SaveSelectionsForm(db *sql.DB, c *gin.Context, selectionID int, selectionLink, selectionName string, winnersOnly bool, date string) error {

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

	if !winnersOnly {
		if selection_id == 0 {
			// Scrape and clean the data
			selectionsForm, err := GetAll(selectionLink)
			if err != nil {
				return err
			}
			err = SaveSelectionForm(db, selectionsForm, c, selectionName, selectionID)
			if err != nil {
				return err
			}

		} else {
			// Get the last date of the selection form
			selectionsForm, err := GetLatest(selectionLink, raceDate)

			if err != nil {
				return err
			}

			err = SaveSelectionForm(db, selectionsForm, c, selectionName, selectionID)

			if err != nil {
				return err

			}
		}

	} else {
		// Get the last date of the selection form
		selectionsForm, err := GetWinner(selectionLink, date)

		if err != nil {
			return err
		}
		
		err = SaveWinners(db, selectionsForm, c, selectionName, selectionID, date)

		if err != nil {
			return err

		}

	}

	return nil
}

func SaveSelectionForm(db *sql.DB, selectionsForm []models.SelectionsForm, c *gin.Context, selectionName string, selectionID int) error {

	if len(selectionsForm) == 0 {
		fmt.Println("No selections form data to insert.")
		return nil
	}

	// Start a transaction
	tx, err := db.BeginTx(c, nil)
	if err != nil {
		fmt.Println("Failed to begin transaction:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return err
	}
	fmt.Println("Transaction started")

	for _, selectionForm := range selectionsForm {
		// Processing and conversions (omitted for brevity)
		fmt.Println("Inserting record for:", selectionForm)

		res, err := tx.ExecContext(c, `
        INSERT INTO SelectionsForm (
			selection_name,
			selection_id,
			race_class,
			race_date,
			position,
			rating,
			race_type,
			racecourse,
			distance,
			going,

			sp_odds,
			Age,
			Trainer,
			Sex,
			Sire,
			Dam,
			Owner,
			created_at,
			updated_at
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			selectionName, selectionID, selectionForm.RaceClass, selectionForm.RaceDate, selectionForm.Position,
			selectionForm.Rating, selectionForm.RaceType, selectionForm.Racecourse,
			selectionForm.Distance, selectionForm.Going,
			selectionForm.SPOdds, selectionForm.Age, selectionForm.Trainer,
			selectionForm.Sex, selectionForm.Sire, selectionForm.Dam, selectionForm.Owner,
			time.Now(),
			time.Now())

		if err != nil {
			fmt.Println("Error executing SQL:", err)
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}

		rowsAffected, _ := res.RowsAffected()
		fmt.Println("Rows affected:", rowsAffected)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		fmt.Println("Transaction commit failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return err
	}

	fmt.Println("Transaction committed successfully")

	return nil
}

func GetAll(selectionLink string) ([]models.SelectionsForm, error) {
	c := colly.NewCollector()

	// Slice to store all horse information
	selectionsForm := []models.SelectionsForm{}

	var age, trainer, sex, sire, dam, owner string

	c.OnHTML("table.Header__DataTable-xeaizz-1", func(e *colly.HTMLElement) {
		age = e.ChildText("tr:nth-child(1) td.Header__DataValue-xeaizz-4")
		trainer = e.ChildText("tr:nth-child(2) td.Header__DataValue-xeaizz-4 a")
		sex = e.ChildText("tr:nth-child(3) td.Header__DataValue-xeaizz-4")
		sire = e.ChildText("tr:nth-child(4) td.Header__DataValue-xeaizz-4")
		dam = e.ChildText("tr:nth-child(5) td.Header__DataValue-xeaizz-4")
		owner = e.ChildText("tr:nth-child(6) td.Header__DataValue-xeaizz-4")
	})

	// Now continue with the rest of your code to scrape race data
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

		// Parsing race date to time.Time
		parsedDate, _ := time.Parse("02/01/06", raceDate) // Assuming UK date format

		// Split the date by "/" and add the current year
		dateParts := strings.Split(raceDate, "/")
		raceDate = "20" + dateParts[2] + "-" + dateParts[1] + "-" + dateParts[0]

		// Convert raceDate to time.Time
		parsedRaceDate, _ := time.Parse("2006-01-02", raceDate)
		// sDistance := convertDistance(distance)

		selectionForm := models.SelectionsForm{
			RaceDate:   parsedRaceDate,
			Position:   position,
			Rating:     rating,
			RaceType:   raceType,
			Racecourse: racecourse,
			Distance:   distance,
			Going:      going,
			RaceClass:  class,
			SPOdds:     spOdds,
			RaceURL:    raceLink,
			EventDate:  parsedDate,
			Age:        age,
			Trainer:    trainer,
			Sex:        sex,
			Sire:       sire,
			Dam:        dam,
			Owner:      owner,
			CreatedAt:  time.Now(),
		}

		selectionsForm = append(selectionsForm, selectionForm)
	})

	// Start scraping the URL
	c.Visit("https://www.sportinglife.com" + selectionLink)

	return selectionsForm, nil
}

func GetLatest(selectionLink string, lasRuntDate time.Time) ([]models.SelectionsForm, error) {
	c := colly.NewCollector()

	// Slice to store all horse information
	selectionsForm := []models.SelectionsForm{}

	var age, trainer, sex, sire, dam, owner string

	c.OnHTML("table.Header__DataTable-xeaizz-1", func(e *colly.HTMLElement) {
		age = e.ChildText("tr:nth-child(1) td.Header__DataValue-xeaizz-4")
		trainer = e.ChildText("tr:nth-child(2) td.Header__DataValue-xeaizz-4 a")
		sex = e.ChildText("tr:nth-child(3) td.Header__DataValue-xeaizz-4")
		sire = e.ChildText("tr:nth-child(4) td.Header__DataValue-xeaizz-4")
		dam = e.ChildText("tr:nth-child(5) td.Header__DataValue-xeaizz-4")
		owner = e.ChildText("tr:nth-child(6) td.Header__DataValue-xeaizz-4")
	})

	// Now continue with the rest of your code to scrape other data
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

		// Split the date by "/" and add the current year
		dateParts := strings.Split(raceDate, "/")
		raceDate = "20" + dateParts[2] + "-" + dateParts[1] + "-" + dateParts[0]

		// Convert raceDate to time.Time
		parsedRaceDate, _ := time.Parse("2006-01-02", raceDate)

		if !parsedRaceDate.After(lasRuntDate) {
			return
		}

		// Create a new SelectionsForm object with the scraped data
		selectionForm := models.SelectionsForm{
			RaceDate:   parsedRaceDate,
			Position:   position,
			Rating:     rating,
			RaceType:   raceType,
			Racecourse: racecourse,
			Distance:   distance,
			Going:      going,
			RaceClass:  class,
			SPOdds:     spOdds,
			RaceURL:    raceLink,
			EventDate:  parsedRaceDate,
			Age:        age,
			Trainer:    trainer,
			Sex:        sex,
			Sire:       sire,
			Dam:        dam,
			Owner:      owner,
			CreatedAt:  time.Now(),
		}

		// Append the selection form to the slice
		selectionsForm = append(selectionsForm, selectionForm)

		// Abort the request after processing the first row
		e.Request.Abort()
	})

	// Start scraping the URL
	c.Visit("https://www.sportinglife.com" + selectionLink)

	return selectionsForm, nil
}

func GetWinner(selectionLink string, date string) ([]models.SelectionsForm, error) {
	c := colly.NewCollector()

	// Slice to store all horse information
	selectionsForm := []models.SelectionsForm{}

	// Now continue with the rest of your code to scrape other data
	c.OnHTML("table.FormTable__StyledTable-sc-1xr7jxa-1 tbody tr", func(e *colly.HTMLElement) {
		raceDate := e.ChildText("td:nth-child(1) a")
		position := e.ChildText("td:nth-child(2)")

		

		// Create a new SelectionsForm object with the scraped data
		selectionForm := models.SelectionsForm{
			Position:  position,
			CreatedAt: time.Now(),
		}


		// Split the date by "/" and add the current year
		dateParts := strings.Split(raceDate, "/")
		raceDate = "20" + dateParts[2] + "-" + dateParts[1] + "-" + dateParts[0]



		if raceDate == date {
			selectionsForm = append(selectionsForm, selectionForm)
		}


	})

	// Start scraping the URL
	c.Visit("https://www.sportinglife.com" + selectionLink)

	return selectionsForm, nil
}

func SaveWinners(db *sql.DB, selectionsForm []models.SelectionsForm, c *gin.Context, selectionName string, selectionID int, dateTime string) error {

	if len(selectionsForm) == 0 {
		fmt.Println("No selections form data to insert.")
		return nil
	}

	// Start a transaction
	tx, err := db.BeginTx(c, nil)
	if err != nil {
		fmt.Println("Failed to begin transaction:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return err
	}
	fmt.Println("Transaction started")

	for _, selectionForm := range selectionsForm {
		
		// Processing and conversions (omitted for brevity)
		fmt.Println("Inserting record for:", selectionForm)

		res, err := tx.ExecContext(c, `
										UPDATE EventPredictions
										SET selection_position = ?, updated_at = ?
										WHERE DATE(event_date) = ? and selection_id = ?`,
			selectionForm.Position, time.Now(), dateTime, selectionID)

		if err != nil {
			fmt.Println("Error executing SQL:", err)
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}

		rowsAffected, _ := res.RowsAffected()
		fmt.Println("Rows affected:", rowsAffected)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		fmt.Println("Transaction commit failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return err
	}

	fmt.Println("Transaction committed successfully")

	return nil
}
