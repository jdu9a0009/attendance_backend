package service

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

const baseDir = "statics"

func InArray[T comparable](val T, array []T) bool {
	for _, v := range array {
		if val == v {
			return true
		}
	}
	return false
}

// Upload - uploads file to the specified folder
func Upload(file *multipart.FileHeader, folder string) (path string, err error) {
	targetPath := filepath.Join(baseDir, folder)
	if file == nil {
		return "", nil
	}

	// Updated to include Excel MIME types
	expectedContentType := []string{
		"image/jpeg",
		"image/png",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", // for .xlsx files
		"application/vnd.ms-excel", // for .xls files
	}

	incomeContentType := file.Header.Get("Content-Type")
	if !InArray(incomeContentType, expectedContentType) {
		return "", fmt.Errorf("invalid file type, expected: %v, got: %s", expectedContentType, incomeContentType)
	}

	if _, err := os.Stat(targetPath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(targetPath, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	filepath := filepath.Join(targetPath, time.Now().Format(time.RFC3339)+"-"+file.Filename)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := src.Close(); closeErr != nil {
			log.Println("file upload src.Close() error:", closeErr)
		}
	}()

	out, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			log.Println("file upload out.Close() error:", closeErr)
		}
	}()

	_, err = io.Copy(out, src)
	if err != nil {
		return "", err
	}

	return filepath, nil
}
