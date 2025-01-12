package models

import (
	"api-service/db"
	"api-service/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

type Video struct {
	Uuid       string `json:"uuid" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Hash       string `json:"hash" binding:"required"`
	Format     string `json:"format" binding:"required"`
	UploadedAt string `json:"uploadedAt" binding:"required"`
	Uri        string `json:"uri"`
	Thumbnail  string `json:"thumbnail"`
	FilePath   string `json:"-"` // Excluded from JSON serialization
}

func validateVideoFormat(video *Video) error {
	if video == nil {
		return errors.New("video is nil")
	}

	if video.Uuid == "" || video.Title == "" || video.Hash == "" || video.Format == "" || video.FilePath == "" {
		return errors.New("missing required fields")
	}

	return nil
}

func AddVideo(ctx context.Context, video Video) error {
	if err := validateVideoFormat(&video); err != nil {
		return err
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	// Insert into video table
	videoQuery := `
	INSERT INTO videos (uuid, title, hash, format, file_path, uploaded_at)
	VALUES ($1, $2, $3, $4, $5, DEFAULT)
	`

	_, err = tx.Exec(ctx, videoQuery,
		video.Uuid,
		video.Title,
		video.Hash,
		video.Format,
		video.FilePath,
	)

	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	log.Printf("Video and metadata saved: %v\n", video)
	return nil
}

func DoesHashExists(hash string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT EXISTS (SELECT 1 FROM videos WHERE hash = $1)`
	var exists bool
	err := db.Pool.QueryRow(ctx, query, hash).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func DeleteOldFiles(requiredSpace int64, maxFolderSize int64, retentionDays int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	query := `
		SELECT uuid, file_path
		FROM videos
		WHERE uploaded_at < $1
		ORDER BY uploaded_at ASC
	`

	rows, err := db.Pool.Query(ctx, query, cutoffTime)
	if err != nil {
		return err
	}
	defer rows.Close()

	var filesToDelete []struct {
		UUID     string
		FilePath string
	}
	var totalSizeFreed int64

	for rows.Next() {
		var uuid string
		var filePath string
		if err := rows.Scan(&uuid, &filePath); err != nil {
			return err
		}

		filesToDelete = append(filesToDelete, struct {
			UUID     string
			FilePath string
		}{UUID: uuid, FilePath: filePath})
	}

	for _, file := range filesToDelete {

		info, err := os.Stat(file.FilePath)
		if err != nil {
			log.Printf("Unable to access file %s: %v", file.FilePath, err)
			continue
		}

		fileSize := info.Size()

		// Delete the video file
		if err := os.Remove(file.FilePath); err != nil {
			log.Printf("Error deleting file %s: %v", file.FilePath, err)
			continue
		}

		thumbnailPath := utils.GenerateThumbnailPath(file.UUID)
		if err := os.Remove(thumbnailPath); err != nil {
			log.Printf("Error deleting thumbnail %s: %v", thumbnailPath, err)
		}

		log.Printf("Deleted file: %s (Size: %d bytes) and thumbnail", file.FilePath, fileSize)

		// TODO Count the thumbnail size as well in the freed space but for now we dont care as
		// its not significant
		totalSizeFreed += fileSize

		if totalSizeFreed >= requiredSpace {
			break
		}
	}

	for _, file := range filesToDelete {
		_, err := db.Pool.Exec(ctx, `DELETE FROM videos WHERE uuid = $1`, file.UUID)
		if err != nil {
			log.Printf("Failed to delete record for file %s: %v", file.UUID, err)
		} else {
			log.Printf("Deleted database record for file: %s", file.UUID)
		}
	}

	log.Printf("Total space freed: %d bytes", totalSizeFreed)
	if totalSizeFreed < requiredSpace {
		return fmt.Errorf("unable to free the required space (%d bytes); only %d bytes freed", requiredSpace, totalSizeFreed)
	}

	return nil
}

func GetVideos(ctx context.Context) ([]Video, error) {
	query := `
		SELECT uuid, title, hash, format, uploaded_at
		FROM videos
		ORDER BY uploaded_at DESC
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []Video

	for rows.Next() {
		var video Video
		var uploadedAt time.Time

		if err := rows.Scan(&video.Uuid, &video.Title, &video.Hash, &video.Format, &uploadedAt); err != nil {
			return nil, err
		}

		video.UploadedAt = uploadedAt.Format("2006-01-02 15:04:05")
		video.Uri = utils.BuildVideoURI(video.Uuid, video.Format)
		video.Thumbnail = utils.BuildThumbnailURI(video.Uuid)

		videos = append(videos, video)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

func GetVideoByUUID(ctx context.Context, uuid string) (Video, error) {
	query := `
        SELECT uuid, title, hash, format, file_path, uploaded_at
        FROM videos
        WHERE uuid = $1
        LIMIT 1
    `
	row := db.Pool.QueryRow(ctx, query, uuid)

	var video Video
	var uploadedAt time.Time
	err := row.Scan(&video.Uuid, &video.Title, &video.Hash, &video.Format, &video.FilePath, &uploadedAt)
	if err != nil {
		return Video{}, err
	}

	video.UploadedAt = uploadedAt.Format("2006-01-02 15:04:05")
	video.Uri = utils.BuildVideoURI(video.Uuid, video.Format)
	video.Thumbnail = utils.BuildThumbnailURI(video.Uuid)
	return video, nil
}

func DeleteVideo(ctx context.Context, uuid string) error {
	query := `
        DELETE FROM videos
        WHERE uuid = $1
    `
	_, err := db.Pool.Exec(ctx, query, uuid)
	return err
}
