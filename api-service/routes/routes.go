package routes

import (
	"api-service/config"
	"api-service/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.POST("/publish", controllers.PublishHandler)
		api.DELETE("/delete/:uuid", controllers.DeleteVideoHandler)

		api.GET("/videos", controllers.GetVideosHandler)
		api.GET("/status/:uuid", controllers.GetJobStatusHandler)
	}

	router.Static("/files", config.VideosDir)
}
