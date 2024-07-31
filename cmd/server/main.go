package main

import (
	"log"

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
}
