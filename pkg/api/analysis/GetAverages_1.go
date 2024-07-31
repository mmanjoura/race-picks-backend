package analysis

// import (
// 	"database/sql"
// 	"log"
// 	"math"
// 	"net/http"
// 	"strconv"

// 	"github.com/mmanjoura/race-picks-backend/pkg/database"
// 	"github.com/mmanjoura/race-picks-backend/pkg/models"

// 	"github.com/gin-gonic/gin"
// )

// // / DoCorrelationAnalysis godoc
// // @Summary Do correlation analysis
// // @Description Do correlation analysis
// // @Tags analytics
// // @Accept  json
// // @Produce  json
// // @Param correlationAnalysisParams body models.CorrelationAnalysis true "Correlation analysis parameters"
// // @Success 200 {object} models.CorrelationAnalysis
// // @Router /analytics/DoCorrelationAnalysis [post]
// func GetAverages_2(c *gin.Context) {

// 	db := database.Database.DB
// 	var correlationAnalysisParams models.CorrelationAnalysis

// 	if err := c.ShouldBindJSON(&correlationAnalysisParams); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Execute the query
// 	rows, err := db.Query("SELECT selection_id FROM Races WHERE DATE(event_date) = DATE('now') and event_name = ? ", correlationAnalysisParams.EventName)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Create a slice to hold the event names
// 	var selectionIds []int

// 	// Loop through the rows and append the results to the slice
// 	for rows.Next() {
// 		var selectionId int
// 		if err := rows.Scan(&selectionId); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		selectionIds = append(selectionIds, selectionId)
// 	}

// 	// Check if the correlation Analysis already exists
// 	rows, err = db.Query("select count() from CorrelationAnalysis WHERE DATE(created_at) = DATE('now') and event_name=?", correlationAnalysisParams.EventName)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	defer rows.Close()
// 	// if the correlation analysis already exists do not save it again
// 	var count int
// 	for rows.Next() {
// 		if err := rows.Scan(&count); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 	}
// 	defer rows.Close()

// 	if count == 0 {
// 		err = saveCorrelationAnalysis(selectionIds, db, correlationAnalysisParams.EventName)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return

// 		}
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Correlation analysis saved successfully"})

// }

// func saveCorrelationAnalysis(selectionIds []int, db *sql.DB, eventName string) error {
// 	for _, selectionId := range selectionIds {

// 		rows, err := db.Query(`
// 		WITH horse_data AS (
// 			SELECT
// 				win_lose,
// 				ipmin,
// 				ppmin,
// 				morning_wap AS morning_ppmax,
// 				ipmax,
// 				pp_traded_vol,
// 				selection_name,
// 				selection_id,
// 				SUBSTR(menu_hint, 1, INSTR(menu_hint, ' ') - 1) AS event_name
// 			FROM MarketData
// 			WHERE selection_id = ?
// 		),
// 		stats AS (
// 			SELECT
// 				AVG(CAST(win_lose AS REAL)) AS avg_win_lose,
// 				AVG(ipmin) AS avg_ipmin,
// 				AVG(ppmin) AS avg_ppmin,
// 				AVG(morning_ppmax) AS avg_morning_ppmax,
// 				AVG(ipmax) AS avg_ipmax,
// 				AVG(pp_traded_vol) AS avg_pp_traded_vol,
// 				COUNT(*) AS n
// 			FROM horse_data
// 		)
// 		SELECT
// 			horse_data.*, stats.*
// 		FROM horse_data, stats
// 		`, selectionId)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer rows.Close()

// 		var horseData []models.Analytic
// 		var stats models.Stats

// 		for rows.Next() {
// 			var hd models.Analytic
// 			err := rows.Scan(
// 				&hd.WinLose, &hd.IPMin, &hd.PPMin, &hd.MorningWAP,
// 				&hd.IPMax, &hd.PPTradedVol, &hd.SelectionName, &hd.SelectionID, &hd.EventName,
// 				&stats.AvgWinLose, &stats.AvgIpmin, &stats.AvgPpmin,
// 				&stats.AvgMorningPpmax, &stats.AvgIpmax, &stats.AvgPpTradedVol, &stats.N,
// 			)
// 			if err != nil {
// 				return err
// 			}
// 			horseData = append(horseData, hd)
// 		}

// 		// Convert WinLose from string to float64 and store in WinLoseFloat
// 		for i, hd := range horseData {
// 			winLose, err := strconv.ParseFloat(hd.WinLose, 64)
// 			if err != nil {
// 				return err
// 			}
// 			horseData[i].WinLoseFloat = winLose
// 		}

// 		// Correlation calculations
// 		correlation := func(data []models.Analytic, avgX, avgY float64, n int, getX, getY func(hd models.Analytic) float64) float64 {
// 			var sumXY, sumX2, sumY2 float64
// 			for _, hd := range data {
// 				x := getX(hd) - avgX
// 				y := getY(hd) - avgY
// 				sumXY += x * y
// 				sumX2 += x * x
// 				sumY2 += y * y
// 			}
// 			return sumXY / (math.Sqrt(sumX2) * math.Sqrt(sumY2))
// 		}

// 		winLoseIpmin := correlation(horseData, stats.AvgWinLose, stats.AvgIpmin, stats.N, func(hd models.Analytic) float64 { return hd.WinLoseFloat }, func(hd models.Analytic) float64 { return hd.IPMin })
// 		winLosePpmin := correlation(horseData, stats.AvgWinLose, stats.AvgPpmin, stats.N, func(hd models.Analytic) float64 { return hd.WinLoseFloat }, func(hd models.Analytic) float64 { return hd.PPMin })
// 		winLoseMorningPpmax := correlation(horseData, stats.AvgWinLose, stats.AvgMorningPpmax, stats.N, func(hd models.Analytic) float64 { return hd.WinLoseFloat }, func(hd models.Analytic) float64 { return hd.MorningWAP })
// 		winLoseIpmax := correlation(horseData, stats.AvgWinLose, stats.AvgIpmax, stats.N, func(hd models.Analytic) float64 { return hd.WinLoseFloat }, func(hd models.Analytic) float64 { return hd.IPMax })
// 		winLosePptraded := correlation(horseData, stats.AvgWinLose, stats.AvgPpTradedVol, stats.N, func(hd models.Analytic) float64 { return hd.WinLoseFloat }, func(hd models.Analytic) float64 { return hd.PPTradedVol })

// 		// Insert into CorrelationAnalysis
// 		_, err = db.Exec(`
// 			INSERT INTO CorrelationAnalysis (event_name, selection_name, selection_id, win_lose_ipmin, win_lose_ppmin, win_lose_morning_ppmax, win_lose_ipmax, win_lose_pptraded)
// 			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
// 			horseData[0].EventName, horseData[0].SelectionName, horseData[0].SelectionID, winLoseIpmin, winLosePpmin, winLoseMorningPpmax, winLoseIpmax, winLosePptraded,
// 		)

// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
