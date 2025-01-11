package services

import (
	"api-service/api/models"
	"api-service/config"
	"api-service/utils"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func HandlePublication(c *gin.Context, title string) (*models.Video, int, error) {

	// Required fields
	var videoUuid string = uuid.New().String()
	var err error

	fileHeader, _ := c.FormFile("file")
	extension, err := utils.CheckMimeType(fileHeader)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("invalid file MIME type")
	}

	var path string = filepath.Join(config.VideosDir, videoUuid+extension)
	if err := c.SaveUploadedFile(fileHeader, path); err != nil {
		log.Printf("failed to save file: %v", err)
		return nil, http.StatusInternalServerError, errors.New("failed to save file")
	}

	// avoid the . in the extension
	var format string = extension[1:]
	var hash string = ""
	hash, err = utils.HashFile(path)

	if err != nil {
		os.Remove(path)
		return nil, http.StatusInternalServerError, errors.New("failed to hash file")
	}

	exist, _ := models.DoesHashExists(hash)
	if exist {
		return nil, http.StatusConflict, errors.New("video already exists")
	}

	var video = models.Video{
		Uuid:     videoUuid,
		Title:    title,
		Hash:     hash,
		Format:   format,
		FilePath: path,
	}

	// Save metadata in database
	ctx := c.Request.Context()
	err = models.AddVideo(ctx, video)
	if err != nil {
		if path != "" {
			os.Remove(path)
		}

		log.Printf("Error saving to DB: %v", err.Error())
		return nil, http.StatusInternalServerError, errors.New("failed to save video data")
	}

	return &video, http.StatusCreated, nil
}
