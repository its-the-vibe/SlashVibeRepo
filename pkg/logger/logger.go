package logger

import (
	"log"
	"strings"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type Logger struct {
	level Level
}

func (l *Logger) Level() Level {
	return l.level
}

func New(levelName string) *Logger {
	var level Level
	switch strings.ToLower(levelName) {
	case "debug":
		level = DebugLevel
	case "info":
		level = InfoLevel
	case "warn", "warning":
		level = WarnLevel
	case "error":
		level = ErrorLevel
	default:
		level = InfoLevel
	}
	return &Logger{level: level}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DebugLevel {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		log.Printf("[INFO] "+format, v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= WarnLevel {
		log.Printf("[WARN] "+format, v...)
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ErrorLevel {
		log.Printf("[ERROR] "+format, v...)
	}
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	log.Fatalf("[FATAL] "+format, v...)
}
