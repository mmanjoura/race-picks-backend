package common

import (
	"strconv"
	"strings"
)

type Selection struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	EventName string `json:"event_name"`
	EventDate string `json:"event_date"`
	EventTime string `json:"event_time"`
	Distance  string `json:"event_distance"`
	Odds      string `json:"odds"`
}

func ParseDistance(distanceStr string) int {
	// Example formats:
	// "2m 4f 97y" -> 4577 yards
	// "2m 3f 210y" -> 4482 yards
	// "1m 6f" -> 3520 yards
	// "6f" -> 1320 yards
	var yards int

	// Split into components
	parts := strings.Fields(distanceStr)

	for _, part := range parts {
		if strings.HasSuffix(part, "m") {
			// Convert miles to yards (1 mile = 1760 yards)
			miles, _ := strconv.Atoi(strings.TrimSuffix(part, "m"))
			yards += miles * 1760
		} else if strings.HasSuffix(part, "f") {
			// Convert furlongs to yards (1 furlong = 220 yards)
			furlongs, _ := strconv.Atoi(strings.TrimSuffix(part, "f"))
			yards += furlongs * 220
		} else if strings.HasSuffix(part, "y") {
			// Convert yards directly
			additionalYards, _ := strconv.Atoi(strings.TrimSuffix(part, "y"))
			yards += additionalYards
		}
	}

	return yards
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
