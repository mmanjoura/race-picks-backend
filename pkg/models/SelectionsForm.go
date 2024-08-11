package models

import (
	"time"
)

type SelectionsForm struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	SelectionID int       `json:"selection_id"`
	RaceDate    time.Time `json:"race_date"`
	Position    string    `json:"position"`
	Rating      int       `json:"rating"`
	RaceType    string    `json:"race_type"`
	Racecourse  string    `json:"racecourse"`
	Distance    string    `json:"distance"`
	Going       string    `json:"going"`
	Class        int       `json:"class"`
	SPOdds      string    `json:"sp_odds"`
	RaceURL     string    `json:"race_url"`
	EventDate   time.Time `json:"event_date"`
	CreatedAt   time.Time `json:"created_at"`
}
