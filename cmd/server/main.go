package main

import (
	"log"
	"time"

	"github.com/mmanjoura/race-picks-backend/pkg/api"
	"github.com/mmanjoura/race-picks-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {

	database.ConnectDatabase()
	config := database.Database.Config

	//gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.DebugMode)
	r := api.InitRouter()
	if err := r.Run(config["PORT"]); err != nil {
		log.Fatal(err)
	}
	eventsData()
}

func eventsData() {
	// Start the ticker in a separate goroutine
	go func() {
		for {
			// Get the current time
			now := time.Now()
			if now.Hour() == 1 {
				err := api.Scrape()
				if err != nil {
					log.Println("Error scraping data: ", err)
				}
			}

			// Check if the current hour is between 13:00 and 19:00
			if now.Hour() >= 13 && now.Hour() <= 20 {
				err := api.Scrape()
				if err != nil {
					log.Println("Error scraping data: ", err)
				}
			}

			// Calculate the next hour's start time
			nextHour := now.Truncate(time.Hour).Add(time.Hour)

			// Calculate the duration until the next hour
			durationUntilNextHour := time.Until(nextHour)

			// Sleep until the next hour
			time.Sleep(durationUntilNextHour)
		}
	}()

	// Keep the main function running
	select {} // Blocks forever
}
