package utils

import (
	"api-service/config"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// DetectAndValidateMimeType detects the MIME type of a file from an io.Reader and validates it against allowed MIME types.
func DetectAndValidateMimeType(reader io.Reader) (string, error) {
	buffer := make([]byte, 512)

	_, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for MIME type detection: %w", err)
	}

	// Detect MIME type
	mimeType := http.DetectContentType(buffer)

	// Validate MIME type
	extension, allowed := config.AllowedMimeTypes[mimeType]
	if !allowed {
		return "", fmt.Errorf("file MIME type %s is not allowed", mimeType)
	}

	return extension, nil
}

// CheckMimeType extracts and validates the MIME type of a multipart file.
func CheckMimeType(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file header: %w", err)
	}
	defer file.Close()

	return DetectAndValidateMimeType(file)
}

// HashFile calculates the SHA-256 hash of a file and encodes it in Base64.
func HashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	return base64.StdEncoding.EncodeToString(hasher.Sum(nil)), nil
}

// HashString calculates the SHA-256 hash of a string and encodes it in Base64.
func HashString(input string) string {
	hasher := sha256.New()
	hasher.Write([]byte(input))
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

func BuildURI(uuid string, format string) string {
	return config.PublicUri + "/files/" + uuid + "." + format
}
