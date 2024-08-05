package analysis

// import (
// 	"net/http"

// 	"github.com/mmanjoura/race-picks-backend/pkg/database"
// 	"github.com/mmanjoura/race-picks-backend/pkg/models"

// 	"github.com/gin-gonic/gin"
// )

// func GetAverages_2(c *gin.Context) {

// 	db := database.Database.DB
// 	var modelData []models.AnalysisData

// 	// Execute the query
// 	rows, err := db.Query(`
// 						WITH RankedHorses AS (
// 							SELECT
// 								event_id,
// 								event_name,
// 								selection_id,
// 								selection_name,
// 								bsp,
// 								ppwap,
// 								morning_wap,
// 								ppmin,
// 								ipmin,
// 								morning_traded_vol,
// 								pp_traded_vol,
// 								ip_traded_vol,
// 								RANK() OVER (PARTITION BY event_id ORDER BY bsp ASC, ppwap ASC, ppmin ASC, ipmin ASC, 
// 									morning_traded_vol DESC, pp_traded_vol DESC, ip_traded_vol DESC) AS horse_rank
// 							FROM
// 								MarketData
// 							WHERE  selection_id IN (
// 								select selection_id from TodayRunners where event_name = 'Goodwood'
// 								)

// 						)
// 						SELECT
// 							event_id,
// 							event_name,
// 							selection_id,
// 							selection_name,
// 							COALESCE(AVG(bsp) FILTER (WHERE bsp IS NOT NULL), 0.0) AS avg_bsp, 
// 							COALESCE(AVG(ppwap) FILTER (WHERE ppwap IS NOT NULL), 0.0) AS avg_ppwap, 
// 							COALESCE(AVG(morning_wap) FILTER (WHERE morning_wap IS NOT NULL), 0.0) AS avg_morning_wap, 				
// 							COALESCE(AVG(ppmin) FILTER (WHERE ppmin IS NOT NULL), 0.0) AS avg_ppmin, 
// 							COALESCE(AVG(ipmin) FILTER (WHERE ipmin IS NOT NULL), 0.0) AS avg_ipmin, 
// 							COALESCE(AVG(morning_traded_vol) FILTER (WHERE morning_traded_vol IS NOT NULL), 0.0) AS avg_morning_traded_vol, 
// 							COALESCE(AVG(pp_traded_vol) FILTER (WHERE pp_traded_vol IS NOT NULL), 0.0) AS avg_pp_traded_vol, 
// 							COALESCE(AVG(ip_traded_vol) FILTER (WHERE ip_traded_vol IS NOT NULL), 0.0) AS avg_ip_traded_vol 
// 						FROM
// 							RankedHorses
// 						WHERE
// 							horse_rank = 1 and selection_id = 12085421
// 							order by selection_name`, 1)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var data models.AnalysisData
// 		if err := rows.Scan(

// 			&data.EventID,
// 			&data.EventName,
// 			&data.EventDT,
// 			&data.SelectionID,
// 			&data.SelectionName,
// 			&data.BSP,
// 			&data.PPWAP,
// 			&data.MorningWAP,
// 			&data.PPMin,
// 			&data.IPMin,
// 			&data.MorningTradedVol,
// 			&data.PPTradedVol,
// 			&data.IPTradedVol,
// 		); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		modelData = append(modelData, data)

// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Correlation analysis saved successfully"})

// }
