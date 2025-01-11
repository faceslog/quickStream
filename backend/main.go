package main

import (
	"api-service/api/routes"
	"api-service/config"
	"api-service/db"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	config.SetupEnv()

	db.Init()
	defer db.Close() // Ensure DB connection is closed on exit

	port := config.Port
	host := config.Host

	router := gin.Default()
	routes.RegisterRoutes(router)

	log.Printf("Storing videos in %s for %d days", config.VideosDir, config.RetentionDays)
	router.Run(host + ":" + port)
}
