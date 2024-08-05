package preparation

// import (
// 	"database/sql"
// 	"log"

// 	"net/http"

// 	"github.com/mmanjoura/race-picks-backend/pkg/database"
// 	"github.com/mmanjoura/race-picks-backend/pkg/models"

// 	"github.com/gin-gonic/gin"
// )

// type Events struct {
// 	EventName string `json:"event_name"`
// 	EventID   int    `json:"event_id"`
// }


// func SaveAnalysisData(c *gin.Context) {

// 	db := database.Database.DB

// 	// Execute the query
// 	rows, err := db.Query("SELECT selection_id, event_name FROM TodayRunners WHERE DATE(event_date) = DATE('now')")
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	// if there are no runners for today return an error
// 	if !rows.Next() {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "No runners for today"})
// 		return
// 	}

// 	var evnents []Events

// 	// Loop through the rows and append the results to the slice
// 	for rows.Next() {
// 		event := Events{}
// 		if err := rows.Scan(&event.EventID, &event.EventName); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		evnents = append(evnents, event)
// 	}

// 	// Check if the correlation Analysis already exists
// 	rows, err = db.Query("select count() from AnalysisData WHERE DATE(created_at) = DATE('now')")
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
// 		err = saveAnalysisData(evnents, db)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return

// 		}
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Correlation analysis saved successfully"})

// }

// func saveAnalysisData(events []Events, db *sql.DB) error {
// 	var analysisDataList []models.AnalysisData
// 	for _, event := range events {
// 		rows, err := db.Query(`
// 			SELECT
// 				event_id,
// 				menu_hint,
// 				event_name,
// 				event_dt,
// 				selection_id,
// 				selection_name,
// 				win_lose,
// 				bsp,
// 				ppwap,
// 				morning_wap,
// 				ppmax,
// 				ppmin,
// 				ipmax,
// 				ipmin,
// 				morning_traded_vol,
// 				pp_traded_vol,
// 				ip_traded_vol
// 			FROM MarketData
// 			WHERE selection_id = ?`, event.EventID)

// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer rows.Close()

// 		for rows.Next() {
// 			var analysisData models.AnalysisData
// 			err := rows.Scan(
// 				&analysisData.EventID,
// 				&analysisData.MenuHint,
// 				&analysisData.EventName,
// 				&analysisData.EventDT,
// 				&analysisData.SelectionID,
// 				&analysisData.SelectionName,
// 				&analysisData.WinLose,
// 				&analysisData.BSP,
// 				&analysisData.PPWAP,
// 				&analysisData.MorningWAP,
// 				&analysisData.PPMax,
// 				&analysisData.PPMin,
// 				&analysisData.IPMax,
// 				&analysisData.IPMin,
// 				&analysisData.MorningTradedVol,
// 				&analysisData.PPTradedVol,
// 				&analysisData.IPTradedVol,
// 			)
// 			if err != nil {
// 				return err
// 			}
// 			// Override the event_name from the Event struct
// 			analysisData.EventName = event.EventName
// 			analysisDataList = append(analysisDataList, analysisData)
// 		}
// 	}

// 	for _, analysisData := range analysisDataList {
// 		_, err := db.Exec(`
// 			INSERT INTO AnalysisData (
// 				event_id,
// 				menu_hint,
// 				event_name,
// 				event_dt,
// 				selection_id,
// 				selection_name,
// 				win_lose,
// 				bsp,
// 				ppwap,
// 				morning_wap,
// 				ppmax,
// 				ppmin,
// 				ipmax,
// 				ipmin,
// 				morning_traded_vol,
// 				pp_traded_vol,
// 				ip_traded_vol
// 			)
// 			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`,
// 			analysisData.EventID,
// 			analysisData.MenuHint,
// 			analysisData.EventName,
// 			analysisData.EventDT,
// 			analysisData.SelectionID,
// 			analysisData.SelectionName,
// 			analysisData.WinLose,
// 			analysisData.BSP,
// 			analysisData.PPWAP,
// 			analysisData.MorningWAP,
// 			analysisData.PPMax,
// 			analysisData.PPMin,
// 			analysisData.IPMax,
// 			analysisData.IPMin,
// 			analysisData.MorningTradedVol,
// 			analysisData.PPTradedVol,
// 			analysisData.IPTradedVol,
// 		)

// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
