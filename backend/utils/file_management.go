package utils

import (
	"api-service/config"
	"log"
	"os"
	"path/filepath"
	"sort"
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

func DeleteOldFiles(requiredSpace int64) error {
	var files []os.DirEntry
	var currentSize int64

	allFiles, err := os.ReadDir(config.VideosDir)
	if err != nil {
		return err
	}

	sort.Slice(allFiles, func(i, j int) bool {
		infoI, errI := allFiles[i].Info()
		infoJ, errJ := allFiles[j].Info()
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	for _, file := range allFiles {
		info, err := file.Info()
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, file)
			currentSize += info.Size()
		}
	}

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return err
		}

		if currentSize+requiredSpace <= config.MAX_FOLDER_SIZE {
			break
		}

		filePath := filepath.Join(config.VideosDir, file.Name())
		if err := os.Remove(filePath); err != nil {
			return err
		}

		currentSize -= info.Size()
		log.Printf("--> Deleting: %s", filePath)
	}

	return nil
}
