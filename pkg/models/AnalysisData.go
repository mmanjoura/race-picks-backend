package models

import (
	"time"
)
type AnalysisData struct {
	
	ID               int       `json:"id"`
	WinLoseFloat     float64   `json:"win_lose_float"`
	EventID          int       `json:"event_id"`
	MenuHint         string    `json:"menu_hint"`
	EventName        string    `json:"event_name"`
	EventTime		string    `json:"event_time"`
	EventDT          string    `json:"event_dt"`
	SelectionID      int       `json:"selection_id"`
	SelectionName    string    `json:"selection_name"`
	RunCount		 int       `json:"run_count"`
	WinLose          string    `json:"win_lose"`
	BSP              float64   `json:"bsp"`
	PPWAP            float64   `json:"ppwap"`
	MorningWAP       float64   `json:"morning_wap"`
	PPMax            float64   `json:"ppmax"`
	PPMin            float64   `json:"ppmin"`
	IPMax            float64   `json:"ipmax"`
	IPMin            float64   `json:"ipmin"`
	MorningTradedVol float64   `json:"morning_traded_vol"`
	PPTradedVol      float64   `json:"pp_traded_vol"`
	IPTradedVol      float64   `json:"ip_traded_vol"`
	CreateAt         time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

}