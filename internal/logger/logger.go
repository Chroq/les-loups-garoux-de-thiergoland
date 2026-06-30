package logger

import (
	"log"
	"strings"
)

type Level uint8

const (
	LevelProduction Level = iota
	LevelDebug
)

var currentLevel Level = LevelProduction

// SetLevel sets the global log level based on a string value ("debug" or "production").
func SetLevel(levelStr string) {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		currentLevel = LevelDebug
	default:
		currentLevel = LevelProduction
	}
}

// Debugf prints a formatted message only if the log level is set to Debug.
func Debugf(format string, v ...any) {
	if currentLevel >= LevelDebug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// Infof prints a formatted message for production/info events.
func Infof(format string, v ...any) {
	log.Printf("[INFO] "+format, v...)
}

// Errorf prints formatted error messages.
func Errorf(format string, v ...any) {
	log.Printf("[ERROR] "+format, v...)
}
