package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// 5000 << 20 is equivalent to 5GB
var MAX_FILE_SIZE int64 = 5000 << 20

// 10GB
var MAX_FOLDER_SIZE int64 = 10000 << 20

var (
	VideosDir     string
	RetentionDays int
	PublicUri     string
	Host          string
	Port          string
	DbUrl         string
)

var ThumbnailFormat = ".jpg"

// list of supported extensions
var AllowedMimeTypes = map[string]string{
	"video/mp4": ".mp4",
}

// LoadEnv loads environment variables from the .env file.
func SetupEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}
	log.Println(".env file loaded")

	PublicUri = getEnv("PUBLIC_URI", "")
	if PublicUri == "" {
		log.Fatal("PUBLIC_URI is required")
	}

	DbUrl = getEnv("DATABASE_URL", "")
	if DbUrl == "" {
		log.Fatal("DATABASE_URL is required")
	}

	RetentionDays, _ = strconv.Atoi(getEnv("RETENTIONS_DAYS", "7"))

	Host = getEnv("HOST", "localhost")
	Port = getEnv("PORT", "8080")

	VideosDir = getEnv("VIDEO_DIR", "")

	fmt.Println("Creating directories")
	log.Printf("VideosDir: %s\n", VideosDir)

	if err := os.MkdirAll(VideosDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create VideosDir: %v", err)
	}

	log.Println("Directories initialized successfully")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
