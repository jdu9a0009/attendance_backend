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

	expectedContentType := []string{
		"image/jpeg",
		"image/png",
	}

	incomeContentType := file.Header.Values("Content-Type")[0]
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
	defer log.Println("file upload src.Close() error: ", src.Close())

	out, err := os.Create(filepath)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(out, src)
	if err != nil {
		return "", err
	}

	return filepath, nil
}
