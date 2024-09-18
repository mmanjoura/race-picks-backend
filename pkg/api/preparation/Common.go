package preparation

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

func SaveSelectionsForm(db *sql.DB, c *gin.Context, selectionID int, selectionLink, selectionName string, date string) error {

	rows, err := db.Query(`
			select selection_id, event_date  from EventRunners where selection_id = ? order by event_date desc limit 1`, selectionID)
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

	form, err := GetForm(db, selectionLink, selectionID, date)
	if err != nil {
		return err
	}
	for _, fr := range form {
		if err != nil {
			return err
		}

		if fr.RaceDate == raceDate {
			return nil
		}

		err = SaveSelectionForm(db, fr, c, selectionName, selectionID)
		if err != nil {
			return err
		}
	}

	return nil
}

func SaveSelectionForm(db *sql.DB, selectionForm models.SelectionsForm, c *gin.Context, selectionName string, selectionID int) error {

	// Start a transaction
	tx, err := db.BeginTx(c, nil)
	if err != nil {
		fmt.Println("Failed to begin transaction:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return err
	}
	fmt.Println("Transaction started")

	// for _, selectionForm := range selectionsForm {
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
	// }

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

		// if !parsedRaceDate.After(lasRuntDate) {
		// 	return
		// }

		// Get

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

func ConvertDistance(distanceStr string) string {
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
func GetForm(db *sql.DB, selectionLink string, selectionId int, date string) ([]models.SelectionsForm, error) {
	c := colly.NewCollector()

	// Slice to store all horse information
	selectionsForm := []models.SelectionsForm{}

	rows, err := db.Query(`
		select * from selectionsForm where DATE(race_date) = ? and selection_id = ?`, date, selectionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Check the count of rows
	rowsCount := 0
	for rows.Next() {
		rowsCount++
	}

	if !rows.Next() {

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

			parsedDate, _ := time.Parse("02/01/06", raceDate) // Assuming UK date format

			dateParts := strings.Split(raceDate, "/")
			raceDate = "20" + dateParts[2] + "-" + dateParts[1] + "-" + dateParts[0]

			parsedRaceDate, _ := time.Parse("2006-01-02", raceDate)

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

			if raceDate == date {
				selectionsForm = append(selectionsForm, selectionForm)
			}
		})

		// Start scraping the URL
		c.Visit("https://www.sportinglife.com" + selectionLink)
		return selectionsForm, nil

	}

	return nil, nil
}

func GetSelectionForm(selectionLink string) ([]models.SelectionsForm, error) {
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
		// parsedEventDate, _ := time.Parse("2006-01-02", eventDate)

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

	})

	// Start scraping the URL
	c.Visit("https://www.sportinglife.com" + selectionLink)

	return selectionsForm, nil
}

func GePredictionWinners(db *sql.DB, c *gin.Context, selectionID int, eventLink, selectionName string, date string) error {

	rows, err := db.Query(`
			select selection_id, event_date  from EventRunners where selection_id = ? order by event_date desc limit 1`, selectionID)
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

	predictionDetail, err := GetPredictionResult(selectionName, eventLink)
	if err != nil {
		return err
	}
	_ = predictionDetail

	return nil
}

// Function to scrape the position and price based on the horse name
func GetPredictionResult(selectionName, eventLink string) (models.HorseDetails, error) {
	var details models.HorseDetails

	resultLink := strings.Replace(eventLink, "/racecards/", "/results/", 1)
	resultLink = strings.Replace(resultLink, "racecard/", "", 1)

	// Initialize a new collector
	c := colly.NewCollector()

	// Callback for when the horse name is found
	c.OnHTML("div.ResultRunner__StyledResultRunnerWrapper-sc-58kifh-13", func(e *colly.HTMLElement) {
		// Check if the horse name matches the selectionName
		horseName := e.ChildText("div.ResultRunner__StyledHorseName-sc-58kifh-5 a")
		if strings.TrimSpace(horseName) == selectionName {
			// Extract the position
			position := e.ChildText("div[data-test-id='position-no'] span.Ordinal__OrdinalWrapper-sc-3wdkxx-0")

			// Extract the price
			price := e.ChildText("span.BetLink__BetLinkStyle-jgjcm-0 span")

			// Assign values to the HorseDetails struct
			details.Position = strings.TrimSpace(position)
			details.Price = strings.TrimSpace(price)
		}
	})

	// Handle errors in case of any issue
	c.OnError(func(_ *colly.Response, err error) {
		err = fmt.Errorf("Error occurred while scraping: %v", err)
	})

	// Visit the results page
	err := c.Visit("https://www.sportinglife.com" + resultLink)
	if err != nil {
		return models.HorseDetails{}, fmt.Errorf("Failed to visit the URL: %v", err)
	}

	// Check if position and price details were scraped successfully
	if details.Price == "" || details.Position == "" {
		return models.HorseDetails{}, fmt.Errorf("Details not found for horse: %s", selectionName)
	}

	return details, nil
}
