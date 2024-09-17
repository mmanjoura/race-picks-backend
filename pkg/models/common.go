package models

import "time"

// Struct to store the horse's position and price
type HorseDetails struct {
    Position string
    Price    string
	PotentialReturn float64
}

type Selection struct {
	ID      int
	Name    string
	Link    string
	Date    string
	Postion string
}

type HistoricalData struct {
	Date     time.Time
	Position string
	Distance float64
}

type ScoreBreakdown struct {
	EventName     string  `json:"event_name"`
	EventTime     string  `json:"event_time"`
	SelectionName string  `json:"selection_name"`
	Odds          string  `json:"odds"`
	Trainer       string  `json:"trainer"`
	RaceTypeScore float64 `json:"race_type_score"`
	CourseScore   float64 `json:"course_score"`
	DistanceScore float64 `json:"distiance_score"`
	ClassScore    float64 `json:"class_score"`
	AgeScore      float64 `json:"age_score"`
	RatingScore   float64 `json:"rating_score"`
	DateScore     float64 `json:"date-score"`
	PositionScore float64 `json:"position_score"`
}


type EventDate struct  {
	Date string `json:"event_date"`
}
