package preparation

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

/// DownloadMarketData godoc
// @Summary Download the market data
// @Description Download the market data
// @Tags DownloadMarketData
// @Accept  json
// @Produce  json
// @Param startDate query string true "Start Date"
// @Param endDate query string true "End Date"
// @Success 200
// @Router /analytics/download-market-data [get]
func GetMarketData(c *gin.Context) {

	// Get the data
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	err := getFiles(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func getFiles(startDate, endDate string) error {
	// Parse the start and end dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("failed to parse start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("failed to parse end date: %w", err)
	}

	// Loop from startDate to endDate
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		meetingDate := d.Format("02012006")

		// Download dwbfpricesireplace file
		fileUrl := "https://promo.betfair.com/betfairsp/prices/dwbfpricesireplace" + meetingDate + ".csv"
		if err = DownloadFile("./data/dwbfpricesirewin"+meetingDate+".csv", fileUrl); err != nil {
			return fmt.Errorf("failed to download dwbfpricesireplace file for %s: %w", meetingDate, err)
		}
		fmt.Println("Finished downloading result for " + meetingDate + " from Betfair API.")

		// Download dwbfpricesukplace file
		fileUrl = "https://promo.betfair.com/betfairsp/prices/dwbfpricesukplace" + meetingDate + ".csv"
		if err = DownloadFile("./data/dwbfpricesukwin"+meetingDate+".csv", fileUrl); err != nil {
			return fmt.Errorf("failed to download dwbfpricesukplace file for %s: %w", meetingDate, err)
		}
		fmt.Println("Finished downloading result for " + meetingDate + " from Betfair API.")
	}

	return nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
