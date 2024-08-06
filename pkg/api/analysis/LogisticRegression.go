// Explanation:
// Data Extraction:

// The handler first extracts the race details from the TodayRunners table.
// Historical performance data is then extracted from the SelectionsForm table.
// Feature Engineering:

// Features such as the position, rating, and distance from past races are extracted
// for logistic regression.
// The target variable (y) is binary (1 for win, 0 for not win).
// Logistic Regression:

// A logistic regression model is trained using gradient descent. The model tries to predict
// the probability of each horse winning based on its past performance.
// Prediction:

// Once trained, the model predicts the winning probability for each horse,
// and these predictions are returned as the result.
// Sorting:

// The results are sorted by the predicted winning probability.
// This approach assumes basic knowledge of logistic regression and provides a simple gradient descent
// implementation for training. The code can be further refined or extended based on more sophisticated
// requirements or more complex features.
package analysis

// import (
// 	"database/sql"
// 	"encoding/csv"
// 	"log"
// 	"math"
// 	"net/http"
// 	"os"
// 	"strconv"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// 	"github.com/mmanjoura/race-picks-backend/pkg/database"
// 	"gonum.org/v1/gonum/mat"
// 	"gonum.org/v1/gonum/stat"
// )

// type LogisticRegressionResult struct {
// 	SelectionID    int64   `json:"selection_id"`
// 	SelectionName  string  `json:"selection_name"`
// 	EventName      string  `json:"event_name"`
// 	EventTime      string  `json:"event_time"`
// 	PredictedWin   bool    `json:"predicted_win"`
// 	WinProbability float64 `json:"win_probability"`
// }

// // LogisticRegressionPrediction predicts the winner using logistic regression
// func LogisticRegressionPrediction(c *gin.Context) {
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

// 	// Extract features and target variables
// 	var X []float64
// 	var y []float64
// 	for _, selection := range selections {
// 		// Extract features (e.g., historical performance metrics)
// 		features, target := extractFeaturesAndTarget(db, selection.ID, modelparams.Distance)
// 		X = append(X, features...)
// 		y = append(y, target)
// 	}

// 	// Perform logistic regression
// 	if len(X) == 0 || len(y) == 0 {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "No data available for logistic regression"})
// 		return
// 	}

// 	// Train the logistic regression model
// 	coefficients := trainLogisticRegression(X, y)

// 	// Predict the probabilities
// 	var results []LogisticRegressionResult
// 	for _, selection := range selections {
// 		features, _ := extractFeaturesAndTarget(db, selection.ID, modelparams.Distance)
// 		probability := predictLogisticRegression(features, coefficients)
// 		results = append(results, LogisticRegressionResult{
// 			SelectionID:    selection.ID,
// 			SelectionName:  selection.Name,
// 			EventName:      selection.EventName,
// 			EventTime:      selection.EventTime,
// 			PredictedWin:   probability >= 0.5,
// 			WinProbability: probability,
// 		})
// 	}

// 	// Sort results by win probability (optional)
// 	sort.SliceStable(results, func(i, j int) bool {
// 		return results[i].WinProbability > results[j].WinProbability
// 	})

// 	c.JSON(http.StatusOK, gin.H{"data": results})
// }

// func extractFeaturesAndTarget(db *sql.DB, selectionID int64, distance string) ([]float64, float64) {
// 	// Query historical data for this selection
// 	var features []float64
// 	var target float64

// 	rows, err := db.Query(`
// 		SELECT position, rating, distance 
// 		FROM SelectionsForm
// 		WHERE selection_id = ?`,
// 		selectionID)
// 	if err != nil {
// 		log.Println("Error querying historical data:", err)
// 		return nil, 0.0
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
// 				target = 1.0
// 			} else {
// 				target = 0.0
// 			}
// 		}
// 	}

// 	return features, target
// }

// func trainLogisticRegression(X []float64, y []float64) []float64 {
// 	// Convert X and y into matrices
// 	m, n := len(y), len(X)/len(y)
// 	XMat := mat.NewDense(m, n, X)
// 	yMat := mat.NewDense(m, 1, y)

// 	// Add a column of ones to X for the intercept
// 	ones := mat.NewVecDense(m, nil)
// 	for i := 0; i < m; i++ {
// 		ones.SetVec(i, 1.0)
// 	}
// 	XWithIntercept := mat.NewDense(m, n+1, nil)
// 	XWithIntercept.Augment(ones, XMat)

// 	// Initialize coefficients
// 	coefficients := make([]float64, n+1)
// 	coefficientsMat := mat.NewVecDense(n+1, coefficients)

// 	// Gradient descent parameters
// 	alpha := 0.01
// 	numIterations := 10000

// 	for i := 0; i < numIterations; i++ {
// 		// Compute the predictions
// 		predictions := mat.NewVecDense(m, nil)
// 		predictions.MulVec(XWithIntercept, coefficientsMat)
// 		predictions.Apply(sigmoid, predictions)

// 		// Compute the gradient
// 		errorVec := mat.NewVecDense(m, nil)
// 		errorVec.SubVec(predictions, yMat)
// 		gradient := mat.NewVecDense(n+1, nil)
// 		gradient.MulVec(XWithIntercept.T(), errorVec)
// 		gradient.ScaleVec(alpha/float64(m), gradient)

// 		// Update coefficients
// 		coefficientsMat.SubVec(coefficientsMat, gradient)
// 	}

// 	// Return the coefficients
// 	return coefficientsMat.RawVector().Data
// }

// func predictLogisticRegression(features []float64, coefficients []float64) float64 {
// 	// Add the intercept
// 	features = append([]float64{1.0}, features...)

// 	// Compute the linear combination of features and coefficients
// 	z := 0.0
// 	for i := range features {
// 		z += features[i] * coefficients[i]
// 	}

// 	// Apply the sigmoid function to get the probability
// 	return sigmoidScalar(z)
// }

// func sigmoid(i, j int, v float64) float64 {
// 	return 1 / (1 + math.Exp(-v))
// }

// func sigmoidScalar(x float64) float64 {
// 	return 1 / (1 + math.Exp(-x))
// }
