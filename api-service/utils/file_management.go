package utils

import (
	"api-service/config"
	"os"
	"os/exec"
	"path/filepath"
)

func CheckFolderSize(fileSize int64) (bool, error) {
	var currentSize int64

	err := filepath.Walk(config.VideosDir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			currentSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return false, err
	}

	return currentSize+fileSize > config.MAX_FOLDER_SIZE, nil
}

func GenerateThumbnail(inputPath, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-ss", "00:00:01.000", "-vframes", "1", outputPath)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func GenerateThumbnailPath(uuid string) string {
	return filepath.Join(config.VideosDir, uuid+config.ThumbnailFormat)
}
