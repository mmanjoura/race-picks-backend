package models

import (
	"time"
)

type AnalysisData struct {
	ID                  int           `json:"id"`
	SelectionID         int           `json:"selection_id"`
	SelectionName       string        `json:"selection_name"`
	RecoveryDays        int           `json:"recovery_days"`
	NumRuns             int           `json:"num_runs"`
	LastRunDate         string        `json:"last_run_date"`
	Duration            int           `json:"duration"`
	WinCount            int           `json:"win_count"`
	AvgPosition         float64       `json:"avg_position"`
	AvgRating           float64       `json:"avg_rating"`
	AvgDistanceFurlongs float64       `json:"avg_distance_furlongs"`
	AvgOdds             float64       `json:"avg_odds"`
	AllPositions        string        `json:"all_positions"`
	AllDistances        string        `json:"all_distances"`
	AllCources          string        `json:"all_cources"`
	AllRaceDates        string        `json:"all_race_dates"`
	TrendAnalysis       AnalyzeTrends `json:"trend_analysis"`
	CreateAt            time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

// RaceData holds individual race information
type RaceData struct {
	Date     time.Time
	Distance float64
	Position int
	Event    string
}

// AnalyzeTrends holds the analysis result of the horse's performance
type AnalyzeTrends struct {
	BestRaces          []RaceData
	OptimalDistanceMin float64
	OptimalDistanceMax float64
}

type AnalysisDataResponse struct {
	EventCourses []string `json:"event_courses"`
	Selections   []AnalysisData `json:"selections"`
}