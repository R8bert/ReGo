package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

// Logger handles logging to file and optionally stdout
type Logger struct {
	file     *os.File
	logger   *log.Logger
	level    LogLevel
	toStdout bool
}

var defaultLogger *Logger

// InitLogger initializes the default logger
func InitLogger(level LogLevel, toStdout bool) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(configDir, "logs")
	if err := EnsureDir(logDir); err != nil {
		return err
	}

	logFile := filepath.Join(logDir, fmt.Sprintf("rego_%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defaultLogger = &Logger{
		file:     file,
		logger:   log.New(file, "", log.LstdFlags),
		level:    level,
		toStdout: toStdout,
	}

	return nil
}

// CloseLogger closes the log file
func CloseLogger() {
	if defaultLogger != nil && defaultLogger.file != nil {
		defaultLogger.file.Close()
	}
}

func (l *Logger) log(level LogLevel, prefix, msg string) {
	if level < l.level {
		return
	}

	fullMsg := fmt.Sprintf("[%s] %s", prefix, msg)
	l.logger.Println(fullMsg)

	if l.toStdout {
		fmt.Println(fullMsg)
	}
}

// Debug logs a debug message
func Debug(msg string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LogDebug, "DEBUG", fmt.Sprintf(msg, args...))
	}
}

// Info logs an info message
func Info(msg string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LogInfo, "INFO", fmt.Sprintf(msg, args...))
	}
}

// Warn logs a warning message
func Warn(msg string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LogWarn, "WARN", fmt.Sprintf(msg, args...))
	}
}

// Error logs an error message
func Error(msg string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LogError, "ERROR", fmt.Sprintf(msg, args...))
	}
}
