// Explanation:
// Data Preparation:

// The handler starts by extracting the race details from the TodayRunners table.
// The historical performance data for each runner is then extracted from the SelectionsForm table.
// Feature Extraction:

// The features extracted for the Decision Tree include the horse's position, rating, and race distance.
// The target variable is binary, indicating whether the horse won the race (1 for win, 0 for not win).
// Dataset Creation:

// A dataset is created using the extracted features and the target variable.
// This dataset is structured to be compatible with the Decision Tree implementation.
// Decision Tree Training:

// A Decision Tree (ID3) is trained using the dataset created above.
// The tree is trained to predict whether a horse will win a race based on the features.
// Prediction:

// Once the Decision Tree is trained, it predicts the winning probability for each horse.
// The handler then sorts the results by the win probability and returns them.
// Sorting and Response:

// The results are sorted by the predicted winning probability and returned as JSON.
package analysis

// import (
// 	"database/sql"
// 	"log"
// 	"math"
// 	"net/http"
// 	"strconv"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// 	"github.com/mmanjoura/race-picks-backend/pkg/database"
// 	"github.com/sjwhitworth/golearn/base"
// 	"github.com/sjwhitworth/golearn/trees"
// )

// type DecisionTreeResult struct {
// 	SelectionID    int64   `json:"selection_id"`
// 	SelectionName  string  `json:"selection_name"`
// 	EventName      string  `json:"event_name"`
// 	EventTime      string  `json:"event_time"`
// 	PredictedWin   bool    `json:"predicted_win"`
// 	WinProbability float64 `json:"win_probability"`
// }

// // DecisionTreePrediction predicts the winner using Decision Trees
// func DecisionTreePrediction(c *gin.Context) {
// 	db := database.Database.DB
// 	var modelparams Selection

// 	if err := c.ShouldBindJSON(&modelparams); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Get today's runners for the given event_name and event_date
// 	rows, err := db.Query(`
// 		SELECT selection_id, 
// 			   selection_name,
// 			   event_name,
// 			   event_time,
// 			   price
// 		FROM TodayRunners	 
// 		WHERE event_name = ? AND DATE(event_date) = DATE('now') AND event_time = ?`,
// 		modelparams.EventName, modelparams.EventTime)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	var selections []Selection
// 	for rows.Next() {
// 		var selection Selection
// 		if err := rows.Scan(&selection.ID, &selection.Name, &selection.EventName, &selection.EventTime, &selection.Odds); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		selections = append(selections, selection)
// 	}

// 	// Create a dataset for training the Decision Tree
// 	attributes, target, err := createDatasetForDecisionTree(db, selections, modelparams.Distance)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Train the Decision Tree
// 	tree := trees.NewID3DecisionTree(0.6)
// 	err = tree.Fit(attributes, target)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Predict the probabilities
// 	var results []DecisionTreeResult
// 	for _, selection := range selections {
// 		features, _ := extractFeaturesForDecisionTree(db, selection.ID, modelparams.Distance)
// 		instance := base.NewDenseInstance(len(features), 0)
// 		for i, feature := range features {
// 			instance.Set(float64(i), feature)
// 		}

// 		// Predict with the decision tree
// 		prediction, prob := tree.PredictWithProbability(instance)
// 		predictedWin := prediction.String() == "1"
// 		results = append(results, DecisionTreeResult{
// 			SelectionID:    selection.ID,
// 			SelectionName:  selection.Name,
// 			EventName:      selection.EventName,
// 			EventTime:      selection.EventTime,
// 			PredictedWin:   predictedWin,
// 			WinProbability: prob[1],
// 		})
// 	}

// 	// Sort results by win probability (optional)
// 	sort.SliceStable(results, func(i, j int) bool {
// 		return results[i].WinProbability > results[j].WinProbability
// 	})

// 	c.JSON(http.StatusOK, gin.H{"data": results})
// }

// func createDatasetForDecisionTree(db *sql.DB, selections []Selection, distance string) (*base.DenseInstances, base.Attribute, error) {
// 	// Define the attributes (features and target)
// 	attrs := make([]base.Attribute, 0)
// 	attrs = append(attrs, base.NewFloatAttribute("Position"))
// 	attrs = append(attrs, base.NewFloatAttribute("Rating"))
// 	attrs = append(attrs, base.NewFloatAttribute("Distance"))

// 	// Target attribute (binary: win or not)
// 	targetAttr := base.NewCategoricalAttribute()
// 	targetAttr.SetName("Win")

// 	attrs = append(attrs, targetAttr)

// 	// Create a dataset with these attributes
// 	dataset := base.NewDenseInstances()
// 	dataset.ExtendAttrs(attrs)

// 	// Fill the dataset with instances from historical data
// 	for _, selection := range selections {
// 		features, target := extractFeaturesForDecisionTree(db, selection.ID, distance)
// 		instance := base.NewDenseInstance(len(attrs)-1, target)
// 		for i, feature := range features {
// 			instance.Set(float64(i), feature)
// 		}
// 		instance.SetClass(float64(target[0]))
// 		dataset.Add(instance)
// 	}

// 	return dataset, targetAttr, nil
// }

// func extractFeaturesForDecisionTree(db *sql.DB, selectionID int64, distance string) ([]float64, []float64) {
// 	// Query historical data for this selection
// 	var features []float64
// 	var target []float64

// 	rows, err := db.Query(`
// 		SELECT position, rating, distance 
// 		FROM SelectionsForm
// 		WHERE selection_id = ?`,
// 		selectionID)
// 	if err != nil {
// 		log.Println("Error querying historical data:", err)
// 		return nil, nil
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var positionStr, distanceStr string
// 		var rating int
// 		if err := rows.Scan(&positionStr, &rating, &distanceStr); err != nil {
// 			log.Println("Error scanning historical data:", err)
// 			continue
// 		}

// 		// Parse the position string (e.g., "3/11" -> 3)
// 		positionParts := strings.Split(positionStr, "/")
// 		if len(positionParts) > 0 {
// 			position, err := strconv.Atoi(positionParts[0])
// 			if err != nil {
// 				log.Println("Non-numeric position encountered, skipping:", positionParts[0])
// 				continue
// 			}
// 			// Features: Position, Rating, Distance
// 			features = append(features, float64(position), float64(rating), float64(parseDistance(distanceStr)))

// 			// Determine the target variable (win or not)
// 			if position == 1 {
// 				target = append(target, 1.0)
// 			} else {
// 				target = append(target, 0.0)
// 			}
// 		}
// 	}

// 	return features, target
// }

// func parseDistance(distanceStr string) int {
// 	// Example formats:
// 	// "2m 4f 97y" -> 4577 yards
// 	// "2m 3f 210y" -> 4482 yards
// 	// "1m 6f" -> 3520 yards
// 	// "6f" -> 1320 yards
// 	var yards int

// 	// Split into components
// 	parts := strings.Fields(distanceStr)

// 	for _, part := range parts {
// 		if strings.HasSuffix(part, "m") {
// 			// Convert miles to yards (1 mile = 1760 yards)
// 			miles, _ := strconv.Atoi(strings.TrimSuffix(part, "m"))
// 			yards += miles * 1760
// 		} else if strings.HasSuffix(part, "f") {
// 			// Convert furlongs to yards (1 furlong = 220 yards)
// 			furlongs, _ := strconv.Atoi(strings.TrimSuffix(part, "f"))
// 			yards += furlongs * 220
// 		} else if strings.HasSuffix(part, "y") {
// 			// Convert yards directly
// 			additionalYards, _ := strconv.Atoi(strings.TrimSuffix(part, "y"))
// 			yards += additionalYards
// 		}
// 	}

// 	return yards
// }

// func abs(x int) int {
// 	if x < 0 {
// 		return -x
// 	}
// 	return x
// }
