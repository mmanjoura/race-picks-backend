package models

import (
	"time"
)

type AnalysisData struct {
	ID                  int       `json:"id"`
	SelectionID         int       `json:"selection_id"`
	SelectionName       string    `json:"selection_name"`
	RecoveryDays		int       `json:"recovery_days"`
	NumRuns             int       `json:"num_runs"`
	LastRunDate         string    `json:"last_run_date"`
	Duration            int       `json:"duration"`
	WinCount            int       `json:"win_count"`
	AvgPosition         float64   `json:"avg_position"`
	AvgRating           float64   `json:"avg_rating"`
	AvgDistanceFurlongs float64   `json:"avg_distance_furlongs"`
	AvgOdds             float64   `json:"avg_odds"`
	CreateAt            time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
