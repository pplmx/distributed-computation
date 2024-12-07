package internal

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogConfig holds configuration for logging
type LogConfig struct {
	Level       string
	JSONLogging bool
	FilePath    string
	MaxSize     int // megabytes
	MaxBackups  int
	MaxAge      int // days
}

// DefaultLogConfig returns a default log configuration
func DefaultLogConfig() LogConfig {
	return LogConfig{
		Level:       "info",
		JSONLogging: true,
		FilePath:    "./logs/app.log",
		MaxSize:     100, // 100 MB
		MaxBackups:  3,
		MaxAge:      30, // 30 days
	}
}

// init replaces the global log instance with a configured logger
func init() {
	// Use the default log configuration for initialization
	config := DefaultLogConfig()

	// Set up global logger
	log.Logger = ConfigureLogger(config)
}

// ConfigureLogger sets up and returns a configured zerolog logger
func ConfigureLogger(config LogConfig) zerolog.Logger {
	// Set global time format
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Parse log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Create log directory if it doesn't exist
	if config.FilePath != "" {
		logDir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Error().Err(err).Msg("Failed to create log directory")
		}
	}

	// File logger with rotation
	fileLogger := &lumberjack.Logger{
		Filename:   config.FilePath,
		MaxSize:    config.MaxSize, // megabytes
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge, // days
		Compress:   true,
	}

	// Log writers: Console and File
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	var fileWriter zerolog.LevelWriter
	if config.JSONLogging {
		fileWriter = zerolog.MultiLevelWriter(fileLogger)
	} else {
		fileWriter = zerolog.MultiLevelWriter(
			zerolog.ConsoleWriter{
				Out:        fileLogger,
				TimeFormat: time.RFC3339,
			},
		)
	}

	// Combine both console and file outputs
	multiWriter := zerolog.MultiLevelWriter(consoleWriter, fileWriter)

	// Create logger
	logger := zerolog.New(multiWriter).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	return logger
}
