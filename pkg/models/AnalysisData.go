package models

import (
	"time"
)

type AnalysisData struct {
	ID                  int       `json:"id"`
	SelectionID         int       `json:"selection_id"`
	SelectionName       string    `json:"selection_name"`
	NumRuns             int       `json:"num_runs"`
	AvgPosition         float64   `json:"avg_position"`
	AvgRating           float64   `json:"avg_rating"`
	AvgDistanceFurlongs float64   `json:"avg_distance_furlongs"`
	AvgOdds             float64   `json:"avg_odds"`
	CreateAt            time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
