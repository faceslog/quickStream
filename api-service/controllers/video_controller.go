package controllers

import (
	"api-service/config"
	"api-service/models"
	"api-service/utils"
	"api-service/workers"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func PublishHandler(c *gin.Context) {
	title := c.PostForm("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if len(title) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title cannot exceed 255 characters"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil || file == nil || header == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Optional: Check size limits, folder usage, etc. as before...
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

	// 1. Generate UUID
	videoUuid := uuid.New().String()

	// 2. Save file quickly to disk
	extension, err := utils.CheckMimeType(header)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file MIME type"})
		return
	}

	filePath := filepath.Join(config.VideosDir, videoUuid+extension)
	if err := c.SaveUploadedFile(header, filePath); err != nil {
		log.Printf("failed to save file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// 3. Submit job to background
	job := workers.Job{
		VideoUuid: videoUuid,
		Title:     title,
		FilePath:  filePath,
		Extension: extension[1:], // remove dot
	}
	workers.SubmitJob(job)

	// 4. Immediately respond (202 Accepted or 201 Created)
	//    Return a tracking ID so the client can check the status
	c.JSON(http.StatusAccepted, gin.H{
		"uuid":    videoUuid,
		"status":  "processing",
		"message": "File accepted for processing. Check back later.",
	})
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

func DeleteVideoHandler(c *gin.Context) {
	videoUUID := c.Param("uuid")

	// 1. Retrieve the video from DB to get the file path
	video, err := models.GetVideoByUUID(c.Request.Context(), videoUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}

	// Delete the record from the database
	err = models.DeleteVideo(c.Request.Context(), videoUUID)
	if err != nil {
		// If DB deletion fails ?? drama ??
		log.Printf("Error removing video record from DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete the database record"})
		return
	}

	// Remove file from disk
	err = os.Remove(video.FilePath)
	if err != nil {
		// If file is missing or locked, log the error
		log.Printf("Error removing file from disk: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete the file from disk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "video deleted successfully"})
}
