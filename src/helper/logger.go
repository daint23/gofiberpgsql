package helper

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

func LogDebug() *os.File {
	pathDebug := "./src/logs/debug"
	errMkd := os.MkdirAll(pathDebug, 0755)
	if errMkd != nil {
		panic(NewHTTPError(404, errMkd))
	}
	fileName := time.Now().Format("02-01-2006") + ".log"

	file, err := os.OpenFile(filepath.Join(pathDebug, fileName), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	return file
}
