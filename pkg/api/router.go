package api

import (
	"time"

	"github.com/mmanjoura/race-picks-backend/pkg/api/analysis"
	"github.com/mmanjoura/race-picks-backend/pkg/api/preparation"
	"github.com/mmanjoura/race-picks-backend/pkg/api/users"

	"github.com/mmanjoura/race-picks-backend/pkg/auth"
	"github.com/mmanjoura/race-picks-backend/pkg/middleware"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	docs "github.com/mmanjoura/race-picks-backend/cmd/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitRouter initializes the routes for the API
func InitRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(middleware.Cors())
	r.Use(middleware.RateLimiter(rate.Every(1*time.Minute), 600)) // 60 requests per minute
	docs.SwaggerInfo.BasePath = "/api/v1"

	v1 := r.Group("/api/v1")
	{
		v1.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
		// Auth routes
		v1.POST("/auth/login", auth.LoginHandler)
		v1.POST("/auth/register", auth.RegisterHandler)
		v1.POST("/auth/logout", auth.Logout)
		v1.GET("/auth/account", users.Account)

		// preparation routes
		
		// v1.POST("/preparation/SaveSelectionsForm", preparation.SaveSelectionsForm)
		v1.POST("/preparation/ScrapeRacesInfo", preparation.ScrapeRacesInfo)
		v1.POST("/preparation/UpdateSelectionsInfo", preparation.UpdateSelectionsInfo)
		v1.POST("/preparation/SaveMarketData", preparation.SaveMarketData)
		// v1.POST("/preparation/SaveAnalysisData", preparation.SaveAnalysisData)
		v1.GET("/preparation/GetMarketData", preparation.GetMarketData)
		v1.GET("/preparation/GetTodayMeeting", preparation.GetTodayMeeting)
		v1.GET("/preparation/GetMeetingRunners", preparation.GetMeetingRunners)
		v1.GET("/preparation/GetEventNames", preparation.GetEventNames)

		v1.POST("/analysis/MonteCarloSimulation", analysis.MonteCarloSimulation)
		v1.POST("/analysis/RacePicksSimulation", analysis.RacePicksSimulation)
	}

	return r
}
