package main

import (
	"api-service/config"
	"api-service/db"
	"api-service/routes"
	"api-service/workers"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	config.SetupEnv()

	db.Init()
	defer db.Close() // Ensure DB connection is closed on exit

	port := config.Port
	host := config.Host

	workers.StartWorkers(2)

	router := gin.Default()
	router.Use(cors.Default())

	routes.RegisterRoutes(router)

	log.Printf("Storing videos in %s for %d days", config.VideosDir, config.RetentionDays)
	router.Run(host + ":" + port)
}
