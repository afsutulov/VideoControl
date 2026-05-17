package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type logLevel int

const (
	levelDebug logLevel = iota
	levelInfo
	levelWarn
	levelError
)

var currentLogLevel = levelInfo

// initLogging настраивает вывод логов.
func initLogging(cfg LoggingConfig) error {
	switch strings.ToLower(cfg.Level) {
	case "", "info":
		currentLogLevel = levelInfo
	case "debug":
		currentLogLevel = levelDebug
	case "warn", "warning":
		currentLogLevel = levelWarn
	case "error":
		currentLogLevel = levelError
	default:
		return fmt.Errorf("unknown logging.level %q", cfg.Level)
	}

	writers := []io.Writer{}
	if cfg.ToConsole || cfg.File == "" {
		writers = append(writers, os.Stdout)
	}
	if cfg.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.File), 0755); err != nil {
			return err
		}
		f, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
		if err != nil {
			return err
		}
		writers = append(writers, f)
	}
	log.SetOutput(io.MultiWriter(writers...))
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	return nil
}

func logDebug(format string, args ...any) {
	if currentLogLevel <= levelDebug {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func logInfo(format string, args ...any) {
	if currentLogLevel <= levelInfo {
		log.Printf("[INFO] "+format, args...)
	}
}

func logWarn(format string, args ...any) {
	if currentLogLevel <= levelWarn {
		log.Printf("[WARN] "+format, args...)
	}
}

func logError(format string, args ...any) {
	if currentLogLevel <= levelError {
		log.Printf("[ERROR] "+format, args...)
	}
}
