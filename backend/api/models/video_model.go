package models

import (
	"api-service/db"
	"context"
	"errors"
	"log"
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
