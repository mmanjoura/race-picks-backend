package models

import (
	"time"
)

type AnalysisDataResponse struct {
	Parameters   OptimalParameters `json:"parameters"`
	Selections   []AnalysisData    `json:"selections"`
	RaceConditon RaceConditon      `json:"race_condition"`
}

type AnalysisData struct {
	ID            int     `json:"id"`
	SelectionID   int     `json:"selection_id"`
	SelectionName string  `json:"selection_name"`
	SelecionLink  string  `json:"selection_link"`
	EventLink     string  `json:"event_link"`
	EventName     string  `json:"event_name"`
	EventDate     string  `json:"event_date"`
	EventTime     string  `json:"event_time"`
	Position      string  `json:"position"`
	Age           string  `json:"age"`
	Trainer       string  `json:"trainer"`
	Sex           string  `json:"sex"`
	Sire          string  `json:"sire"`
	Dam           string  `json:"dam"`
	Owner         string  `json:"owner"`
	EventClass    string  `json:"event_class"`
	RecoveryDays  float64 `json:"recovery_days"`
	NumRuns       int     `json:"num_runs"`
	LastRunDate   string  `json:"last_run_date"`
	Duration      int     `json:"duration"`
	WinCount      int     `json:"win_count"`
	RaceDate      string  `json:"race_date"`

	AvgDistanceFurlongs float64           `json:"avg_distance_furlongs"`
	AvgOdds             float64           `json:"avg_odds"`
	AllRatings          string            `json:"all_ratings"`
	AllClasses          string            `json:"all_classes"`
	AllRaceTypes        string            `json:"all_race_types"`
	AllPositions        string            `json:"all_positions"`
	AllDistances        string            `json:"all_distances"`
	AllCources          string            `json:"all_cources"`
	AllRaceDates        string            `json:"all_race_dates"`
	TrendAnalysis       AnalyzeTrends     `json:"trend_analysis"`
	Parameters          OptimalParameters `json:"weight_parameters"`
	WinLose             WinLose           `json:"win_lose"`

	NumberOfRunners  string    `json:"number_of_runners"`
	CurrentDistance  float64   `json:"current_distance"`
	TotalScore       float64   `json:"total_score"`
	PreferedDistance float64   `json:"prefered_distance"`
	AvgPosition      float64   `json:"avg_position"`
	AvgRating        float64   `json:"avg_rating"`
	CreateAt         time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
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

type RaceParameters struct {
	ID           int    `json:"id"`
	RaceType     string `json:"race_type"`
	RaceDistance string `json:"race_distance"`
	Handicap     bool   `json:"handicap"`
	RaceClass    string `json:"race_class"`
	EventName    string `json:"event_name"`
	EventDate    string `json:"event_date"`
	EventTime    string `json:"event_time"`
	Positions    string `json:"positions"`
	Years        string `json:"years"`
	Ages         string `json:"ages"`
	BetAmount    string `json:"bet_amount"`
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
	Position      string `json:"position"`
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

type SelectionForm struct {
	ID               int       `json:"id"`
	SelectionName    string    `json:"selection_name"`
	SelectionID      int       `json:"selection_id"`
	RaceDate         time.Time `json:"race_date"` // Use time.Time for date fields
	Position         string    `json:"position"`
	Rating           string    `json:"rating"`
	RaceType         string    `json:"race_type"`
	Racecourse       string    `json:"racecourse"`
	Distance         string    `json:"distance"`
	Going            string    `json:"going"`
	Class            string    `json:"class"`
	SpOdds           string    `json:"sp_odds"` // String format for odds
	Age              string    `json:"age"`
	Trainer          string    `json:"trainer"`
	Sex              string    `json:"sex"`
	Sire             string    `json:"sire"`
	Dam              string    `json:"dam"`
	Owner            string    `json:"owner"`
	AVGPosition      float64   `json:"avg_position"`
	AVGRating        float64   `json:"avg_rating"`
	CurrentEventName string    `json:"current_event_name"`
	CurrentEventDate string    `json:"current_event_date"`
	CurrentEventTime string    `json:"current_event_time"`
	Score            string    `json:"score"`
}

type SelectionResult struct {
	SelectionID       int     `json:"selection_id"`
	EventName         string  `json:"event_name"`
	EventDate         string  `json:"event_date"`
	EventTime         string  `json:"event_time"`
	SelectionName     string  `json:"selection_name"`
	SelectionLink     string  `json:"selection_link"`
	EventClass        string  `json:"event_class"`
	RaceType          string  `json:"race_type"`
	Odds              string  `json:"odds"`
	Trainer           string  `json:"trainer"`
	AvgPosition       float64 `json:"avg_position"`
	AvgRating         float64 `json:"avg_rating"`
	TotalScore        float64 `json:"total_score"`
	Age               string  `json:"age"`
	RunCount          string  `json:"run_count"`
	BetType           string  // New field to store BetType
	SelectionPosition string  // New field to store Selection Position
	PotentialReturn   float64 // New field to store Potential Return

}

type SelectionResultResponse struct {
	SelectionsResult []SelectionResult `json:"selections_result"`
	EventPredictions []EventPrediction `json:"profit_and_loss"`
}

type ProfitAndLoss struct {
	StartingPot   float64 `json:"starting_pot"`
	ProfictLoss   float64 `json:"profit_loss"`
	CleanBetScore float64 `json:"clean_bet_score"`
}

// EventPrediction represents a row from the EventPredictions table.
type EventPrediction struct {
	ID                int       `json:"id"`
	SelectionID       int       `json:"selection_id"`
	SelectionName     string    `json:"selection_name"`
	Odds              float64   `json:"odds"`
	Age               int       `json:"age"`
	CleanBetScore     float64   `json:"clean_bet_score"`
	AveragePosition   float64   `json:"average_position"`
	Position          string    `json:"position"`
	AverageRating     float64   `json:"average_rating"`
	EventName         string    `json:"event_name"`
	EventDate         string    `json:"event_date"`
	RaceDate          string    `json:"race_date"`
	EventTime         string    `json:"event_time"`
	SelectionPosition string    `json:"selection_position"`
	DistanceTolerence float64   `json:"distance_tolerence"`
	NumRunners        string    `json:"num_runners"`
	NumbeRuns         int       `json:"number_runs"`
	PreferredDistance float64   `json:"prefered_distance"`
	CurrentDistance   float64   `json:"current_distance"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type EventPredictionResponse struct {
	Selections  []EventPrediction `json:"selections"`
	TotalBet    float64           `json:"total_bet"`
	TotalReturn float64           `json:"total_return"`
}
