package models

import (
	"api-service/db"
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
	FilePath   string `json:"filePath,omitempty"`
	UploadedAt string `json:"uploadedAt"`
}

func validateVideoFormat(video *Video) error {
	if video == nil {
		return errors.New("video is nil")
	}

	if video.Uuid == "" || video.Title == "" || video.Hash == "" || video.Format == "" || video.FilePath == "" {
		return errors.New("video uuid, uri, format and hash are required")
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
		SELECT file_path
		FROM videos
		WHERE uploaded_at < $1
		ORDER BY uploaded_at ASC
	`

	rows, err := db.Pool.Query(ctx, query, cutoffTime)
	if err != nil {
		return err
	}
	defer rows.Close()

	var filesToDelete []string
	var totalSizeFreed int64

	for rows.Next() {
		var filePath string
		if err := rows.Scan(&filePath); err != nil {
			return err
		}
		filesToDelete = append(filesToDelete, filePath)
	}

	for _, filePath := range filesToDelete {
		info, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Unable to access file %s: %v", filePath, err)
			continue
		}

		fileSize := info.Size()

		if err := os.Remove(filePath); err != nil {
			log.Printf("Error deleting file %s: %v", filePath, err)
			continue
		}

		log.Printf("Deleted file: %s (Size: %d bytes)", filePath, fileSize)
		totalSizeFreed += fileSize

		if totalSizeFreed >= requiredSpace {
			break
		}
	}

	for _, filePath := range filesToDelete {
		_, err := db.Pool.Exec(ctx, `DELETE FROM videos WHERE file_path = $1`, filePath)
		if err != nil {
			log.Printf("Failed to delete record for file %s: %v", filePath, err)
		} else {
			log.Printf("Deleted database record for file: %s", filePath)
		}
	}

	log.Printf("Total space freed: %d bytes", totalSizeFreed)
	if totalSizeFreed < requiredSpace {
		return fmt.Errorf("unable to free the required space (%d bytes); only %d bytes freed", requiredSpace, totalSizeFreed)
	}

	return nil
}
