package models

import (
	"time"
)

type AnalysisDataResponse struct {
	Parameters OptimalParameters `json:"parameters"`
	Selections []AnalysisData    `json:"selections"`
}

type AnalysisData struct {
	ID                  int               `json:"id"`
	SelectionID         int               `json:"selection_id"`
	SelectionName       string            `json:"selection_name"`
	Position            string            `json:"position"`
	Age                 string            `json:"age"`
	Trainer             string            `json:"trainer"`
	Sex                 string            `json:"sex"`
	Sire                string            `json:"sire"`
	Dam                 string            `json:"dam"`
	Owner               string            `json:"owner"`
	RecoveryDays        int               `json:"recovery_days"`
	NumRuns             int               `json:"num_runs"`
	LastRunDate         string            `json:"last_run_date"`
	Duration            int               `json:"duration"`
	WinCount            int               `json:"win_count"`
	AvgPosition         float64           `json:"avg_position"`
	AvgRating           float64           `json:"avg_rating"`
	AvgDistanceFurlongs float64           `json:"avg_distance_furlongs"`
	AvgOdds             float64           `json:"avg_odds"`
	AllPositions        string            `json:"all_positions"`
	AllDistances        string            `json:"all_distances"`
	AllCources          string            `json:"all_cources"`
	AllRaceDates        string            `json:"all_race_dates"`
	TrendAnalysis       AnalyzeTrends     `json:"trend_analysis"`
	Parameters          OptimalParameters `json:"weight_parameters"`
	WinLose             WinLose           `json:"win_lose"`
	CreateAt            time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
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

type OptimalParameters struct {
	ID                           int     `json:"id"`
	RaceType                     string  `json:"race_type"`
	RaceDistance                 float64 `json:"race_distance"`
	Tolerance                    float64 `json:"tolerance"`
	OptimalRecoveryDays          int     `json:"optimal_recovery_days"`
	OptimalNumRuns               int     `json:"optimal_num_runs"`
	OptimalNumYearsInCompetition int     `json:"optimal_num_years_in_competition"`
	OptimalNumWins               int     `json:"optimal_num_wins"`
	OptimalRating                float64 `json:"optimal_rating"`
	OptimalPosition              float64 `json:"optimal_position"`
	OptimalDistance              float64 `json:"optimal_distance"`
	EventName                    string  `json:"event_name"`
	EventDate                    string  `json:"event_date"`
	EventTime                    string  `json:"event_time"`
}

type CurrentHorseData struct {
	Name             string  `json:"name"`
	DaysSinceLastRun int     `json:"days_since_last_run"`
	NumberOfRuns     int     `json:"number_of_runs"`
	YearsRunning     int     `json:"years_running"`
	NumberOfWins     int     `json:"number_of_wins"`
	AverageRating    float64 `json:"average_rating"`
	AveragePosition  float64 `json:"average_position"`
	AverageDistance  float64 `json:"average_distance"`
	AverageOdds      float64 `json:"average_odds"`
}

type WinLose struct {
	SelectionID   int    `json:"selection_id"`
	SelectionName string `json:"selection_name"`
	EventDate     string `json:"event_date"`
	Position      string    `json:"position"`
}
