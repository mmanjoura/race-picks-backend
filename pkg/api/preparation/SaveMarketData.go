package preparation

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mmanjoura/race-picks-backend/pkg/database"
	"github.com/mmanjoura/race-picks-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

// / SaveMarketData godoc
// @Summary Save the market data
// @Description Save the market data
// @Tags SaveMarketData
// @Accept  json
// @Produce  json
// @Success 200 {object} models.MarketData
// @Router /analytics/save-market-data [get]
func SaveMarketData(c *gin.Context) {

	sourcePath := "./data/"
	dir, _ := os.Open(sourcePath)
	db := database.Database.DB

	//Get list of files in dir
	files, _ := dir.Readdir(-1)

	for _, file := range files {
		filePath := sourcePath + file.Name()

		analytics, err := Format(filePath)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, newAnalytic := range analytics {

			_, err = db.ExecContext(c, `
				INSERT INTO MarketData (
					event_id,
					menu_hint,
					event_name,
					event_dt,
					selection_id,
					selection_name,
					win_lose,
					bsp,
					ppwap,
					morning_wap,
					ppmax,
					ppmin,
					ipmax,
					ipmin,
					morning_traded_vol,
					pp_traded_vol,
					ip_traded_vol,
					created_at,
					updated_at
				)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				newAnalytic.EventID,
				newAnalytic.MenuHint,
				newAnalytic.EventName,
				newAnalytic.EventDT,
				newAnalytic.SelectionID,
				newAnalytic.SelectionName,
				newAnalytic.WinLose,
				newAnalytic.BSP,
				newAnalytic.PPWAP,
				newAnalytic.MorningWAP,
				newAnalytic.PPMax,
				newAnalytic.PPMin,
				newAnalytic.IPMax,
				newAnalytic.IPMin,
				newAnalytic.MorningTradedVol,
				newAnalytic.PPTradedVol,
				newAnalytic.IPTradedVol,
				time.Now(),
				time.Now())

		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error storing data": err.Error()})
			return
		}

	}
	// Delete files
	err := deleteFiles(sourcePath, files)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error deleting files": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Data saved successfully"})

}

func Format(filePath string) ([]models.MarketData, error) {

	var results []models.MarketData
	f, _ := os.Open(filePath)
	defer f.Close()

	records, _ := csv.NewReader(f).ReadAll()
	for _, row := range records {
		var result models.MarketData
		// Skip the header
		if strings.ToUpper(row[0]) == "EVENT_ID" {
			continue
		}
		eventID, err := strconv.Atoi(row[0])
		result.EventID = eventID
		const expectedFields = 17
		if len(row) != expectedFields {
			// Log the problematic row and return an error
			log.Printf("Invalid row: expected %d fields, got %d fields. Row: %+v\n", expectedFields, len(row), row)
			return nil, fmt.Errorf("invalid row: expected %d fields, got %d fields", expectedFields, len(row))
		}
		result.MenuHint = row[1]
		result.EventName = row[2]
		result.EventDT = row[3]
		selectionID, err := strconv.Atoi(row[4])
		result.SelectionID = selectionID
		result.SelectionName = row[5]
		result.WinLose = row[6]
		bsp, err := strconv.ParseFloat(row[7], 64)
		result.BSP = bsp
		ppwap, err := strconv.ParseFloat(row[8], 64)
		result.PPWAP = ppwap
		morningWAP, err := strconv.ParseFloat(row[9], 64)
		result.MorningWAP = morningWAP
		ppMax, err := strconv.ParseFloat(row[10], 64)
		result.PPMax = ppMax
		ppMin, err := strconv.ParseFloat(row[11], 64)
		result.PPMin = ppMin
		ipMax, err := strconv.ParseFloat(row[12], 64)
		result.IPMax = ipMax
		ipMin, err := strconv.ParseFloat(row[13], 64)
		result.IPMin = ipMin
		morningTradedVol, err := strconv.ParseFloat(row[14], 64)
		result.MorningTradedVol = morningTradedVol
		ppTradedVol, err := strconv.ParseFloat(row[15], 64)
		result.PPTradedVol = ppTradedVol
		ipTradedVol, err := strconv.ParseFloat(row[16], 64)
		result.IPTradedVol = ipTradedVol
		if err != nil {
			// Optionally log the error along with the row for debugging
			log.Printf("Error parsing row: %+v, Error: %v\n", row, err)
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// deleteFiles takes a slice of fs.FileInfo and deletes each file
func deleteFiles(filePath string, fileInfos []fs.FileInfo) error {
	for _, fileInfo := range fileInfos {
		// Get the file path from fileInfo
		file := fileInfo.Name()

		// Attempt to remove the file
		err := os.Remove(filePath + file)
		if err != nil {
			// Return an error if something goes wrong
			return fmt.Errorf("error deleting file %s: %w", file, err)
		}
		fmt.Println("Successfully deleted:", file)
	}
	return nil
}
