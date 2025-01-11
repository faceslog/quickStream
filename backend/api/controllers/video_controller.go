package controllers

import (
	"api-service/api/models"
	"api-service/api/services"
	"api-service/config"
	"api-service/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PublishHandler(c *gin.Context) {

	title := c.PostForm("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if len(title) > 255 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "title cannot exceed 255 characters"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil || file == nil || header == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	if header.Size > config.MAX_FILE_SIZE {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file size exceeds limit"})
		return
	}

	folderExceeds, err := utils.CheckFolderSize(header.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check folder size"})
		return
	}

	if folderExceeds {
		err := models.DeleteOldFiles(header.Size, config.MAX_FOLDER_SIZE, config.RetentionDays)
		if err != nil {
			log.Printf("Error freeing up space %v: ", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to free up space"})
			return
		}
	}

	video, status, err := services.HandlePublication(c, title)
	if err != nil {

		if status == http.StatusInternalServerError {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Oops, something went wrong with that file."})
			log.Printf("Failed to handle file publication: %v", err.Error())
			return
		}

		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	var uri string = utils.BuildURI(video.Uuid, video.Format)
	c.JSON(http.StatusCreated, gin.H{"uuid": video.Uuid, "uri": uri})
}

func GetVideosHandler(c *gin.Context) {
	ctx := c.Request.Context()

	videos, err := models.GetVideos(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch videos"})
		return
	}

	c.JSON(http.StatusOK, videos)
}
