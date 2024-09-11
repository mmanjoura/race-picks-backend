package common

import (
	"strconv"
	"strings"
)

type Selection struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	EventName       string `json:"event_name"`
	EventDate       string `json:"event_date"`
	EventTime       string `json:"event_time"`
	Odds            string `json:"odds"`
	Position		string `json:"position"`
	RaceCategory    string `json:"race_category"`
	RaceDistance    string `json:"race_distance"`
	TrackCondition  string `json:"track_condition"`
	NumberOfRunners string `json:"number_of_runners"`
	RaceTrack       string `json:"race_track"`
	RaceClass       string `json:"race_class"`
	Link            string `json:"link"`
	EventLink	   string `json:"event_link"`
}

// Helper function to parse race distance considering miles, furlongs, and yards
func ParseDistance(dist string) float64 {
	var totalFurlongs float64

	// Split the distance string into components (miles, furlongs, yards)
	parts := strings.Split(dist, " ")

	for _, part := range parts {
		if strings.HasSuffix(part, "m") { // Handle miles
			miles, _ := strconv.ParseFloat(strings.TrimSuffix(part, "m"), 64)
			totalFurlongs += miles * 8.0 // 1 mile = 8 furlongs
		} else if strings.HasSuffix(part, "f") { // Handle furlongs
			furlongs, _ := strconv.ParseFloat(strings.TrimSuffix(part, "f"), 64)
			totalFurlongs += furlongs
		} else if strings.HasSuffix(part, "y") { // Handle yards
			yards, _ := strconv.ParseFloat(strings.TrimSuffix(part, "y"), 64)
			totalFurlongs += yards / 220.0 // 1 furlong = 220 yards
		}
	}

	return totalFurlongs
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Helper function to parse odds string to a numeric value (e.g., decimal odds)
func ParseOdds(oddsStr string) float64 {
	// Implement your logic to convert odds from string to a numeric value
	// For example, convert odds from fractional to decimal format
	if strings.Contains(oddsStr, "/") {
		parts := strings.Split(oddsStr, "/")
		if len(parts) == 2 {
			numerator, err1 := strconv.ParseFloat(parts[0], 64)
			denominator, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil && denominator != 0 {
				return numerator/denominator + 1
			}
		}
	}
	return 0.0
}
