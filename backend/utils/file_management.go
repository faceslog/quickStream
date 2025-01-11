package utils

import (
	"api-service/config"
	"os"
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
