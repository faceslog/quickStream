package routes

import (
	"api-service/api/controllers"
	"api-service/config"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.POST("/publish", controllers.PublishHandler)
		api.GET("/videos", controllers.GetVideosHandler)
	}

	router.Static("/videos", config.VideosDir)
}
