package analysis

// import (
// 	"fmt"
// 	"math"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"github.com/mmanjoura/race-picks-backend/pkg/database"
// 	"github.com/mmanjoura/race-picks-backend/pkg/models"
// 	"github.com/tealeg/xlsx"
// )



// func GetAverages_0(c *gin.Context) {
// 	db := database.Database.DB
// 	var modelparams models.AnalysisData

// 	if err := c.ShouldBindJSON(&modelparams); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Get today's runners for the given event_name and event_date
// 	rows, err := db.Query("SELECT selection_id, selection_name FROM TodayRunners WHERE event_name = ? AND DATE(event_date) = DATE('now') AND event_time = ?", modelparams.EventName, modelparams.EventTime)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	// Create a slice to hold the selections
// 	var selections []Selection
// 	for rows.Next() {
// 		var selection Selection
// 		err := rows.Scan(&selection.ID, &selection.Name)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		selections = append(selections, selection)
// 	}

// 	// Calculate the average of the BSP and other fields
// 	var eventResults []models.AnalysisData
// 	for _, selection := range selections {

// 		var selectionData models.AnalysisData
// 		var eventResult models.AnalysisData
// 		selectionData.SelectionName = selection.Name
// 		err := db.QueryRow(`
// 				SELECT 
// 					selection_name,
// 					COUNT(*) AS count,
// 					COALESCE(AVG(bsp) FILTER (WHERE bsp IS NOT NULL), 0.0) AS avg_bsp, 
// 					COALESCE(AVG(ppwap) FILTER (WHERE ppwap IS NOT NULL), 0.0) AS avg_ppwap, 
// 					COALESCE(AVG(morning_wap) FILTER (WHERE morning_wap IS NOT NULL), 0.0) AS avg_morning_wap, 
// 					COALESCE(AVG(ppmax) FILTER (WHERE ppmax IS NOT NULL), 0.0) AS avg_ppmax, 
// 					COALESCE(AVG(ppmin) FILTER (WHERE ppmin IS NOT NULL), 0.0) AS avg_ppmin, 
// 					COALESCE(AVG(ipmax) FILTER (WHERE ipmax IS NOT NULL), 0.0) AS avg_ipmax, 
// 					COALESCE(AVG(ipmin) FILTER (WHERE ipmin IS NOT NULL), 0.0) AS avg_ipmin, 
// 					COALESCE(AVG(morning_traded_vol) FILTER (WHERE morning_traded_vol IS NOT NULL), 0.0) AS avg_morning_traded_vol, 
// 					COALESCE(AVG(pp_traded_vol) FILTER (WHERE pp_traded_vol IS NOT NULL), 0.0) AS avg_pp_traded_vol, 
// 					COALESCE(AVG(ip_traded_vol) FILTER (WHERE ip_traded_vol IS NOT NULL), 0.0) AS avg_ip_traded_vol 
// 				FROM 
// 					MarketData 
// 				WHERE 
// 					selection_id = ?
// 				GROUP BY
// 					selection_name
// 				HAVING
// 			COUNT(*) <= 7`, selection.ID).Scan(
// 			&selectionData.SelectionName,
// 			&selectionData.RunCount,
// 			&selectionData.BSP,
// 			&selectionData.PPWAP,
// 			&selectionData.MorningWAP,
// 			&selectionData.PPMax,
// 			&selectionData.PPMin,
// 			&selectionData.IPMax,
// 			&selectionData.IPMin,
// 			&selectionData.MorningTradedVol,
// 			&selectionData.PPTradedVol,
// 			&selectionData.IPTradedVol)

// 		if err != nil {
// 			if err.Error() == "sql: no rows in result set" {
// 				continue
// 			}
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

// 		if IsCloseEnough(modelparams.BSP, selectionData.BSP, 3) &&
// 			IsCloseEnough(modelparams.PPWAP, selectionData.PPWAP, 3) &&
// 			IsCloseEnough(modelparams.MorningWAP, selectionData.MorningWAP, 3) &&
// 			IsCloseEnough(modelparams.PPMin, selectionData.PPMin, 3) {

// 			eventResult.SelectionName = selectionData.SelectionName
// 			eventResult.BSP = selectionData.BSP
// 			eventResult.PPWAP = selectionData.PPWAP
// 			eventResult.MorningWAP = selectionData.MorningWAP
// 			eventResult.PPMax = selectionData.PPMax
// 			eventResult.PPMin = selectionData.PPMin
// 			eventResult.IPMax = selectionData.IPMax
// 			eventResult.IPMin = selectionData.IPMin
// 			eventResult.IPTradedVol = selectionData.IPTradedVol

// 			eventResults = append(eventResults, eventResult)

// 		}

// 	}

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Horse information saved successfully", "data": eventResults})
// }

// // IsCloseEnough checks if a and b are close enough within the given tolerance
// func IsCloseEnough(a, b, tolerance float64) bool {
// 	return math.Abs(a-b) <= tolerance
// }

// // IsCloseEnough checks if a and b are close enough within the given tolerance
// func getTolerence(a, b float64) float64 {
// 	return float64(math.Abs(a - b))
// }

// func writeToSpreadsheet(selectionsData []models.AnalysisData, filename string) error {
// 	file := xlsx.NewFile()
// 	sheet, err := file.AddSheet("Sheet1")
// 	if err != nil {
// 		return err
// 	}

// 	// Write headers
// 	row := sheet.AddRow()
// 	row.AddCell().Value = "SelectionName"
// 	row.AddCell().Value = "BSP"
// 	row.AddCell().Value = "PPWAP"
// 	row.AddCell().Value = "MorningWAP"
// 	row.AddCell().Value = "PPMax"
// 	row.AddCell().Value = "PPMin"
// 	row.AddCell().Value = "IPMax"
// 	row.AddCell().Value = "IPMin"
// 	row.AddCell().Value = "MorningTradedVol"
// 	row.AddCell().Value = "PPTradedVol"
// 	row.AddCell().Value = "IPTradedVol"
// 	row.AddCell().Value = "WinLose"

// 	// Write selectionData
// 	for _, selection := range selectionsData {
// 		row = sheet.AddRow()
// 		row.AddCell().Value = selection.SelectionName
// 		row.AddCell().Value = fmt.Sprintf("%f", selection.BSP)
// 		row.AddCell().Value = fmt.Sprintf("%f", selection.PPWAP)
// 		row.AddCell().Value = fmt.Sprintf("%f", selection.MorningWAP)
// 		row.AddCell().Value = fmt.Sprintf("%f", selection.PPMax)
// 		row.AddCell().Value = fmt.Sprintf("%f", selection.PPMin)
// 		row.AddCell().Value = fmt.Sprintf("%f", selection.IPMax)
// 		row.AddCell().Value = fmt.Sprintf("%f", selection.IPMin)
// 		row.AddCell().Value = fmt.Sprintf("%f", selection.MorningTradedVol)
// 		row.AddCell().Value = fmt.Sprintf("%f", selection.PPTradedVol)
// 		row.AddCell().Value = fmt.Sprintf("%f", selection.IPTradedVol)
// 		row.AddCell().Value = selection.WinLose
// 	}

// 	return file.Save(filename)
// }
