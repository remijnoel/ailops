package internal

import log "github.com/sirupsen/logrus"
import (
	"os"
	"strings"
)

func init() {
	// Set log level based on environment variable
	if level, ok := os.LookupEnv("AILOPS_LOG_LEVEL"); ok {
		// Normalize the log level to lowercase
		level = strings.ToLower(level)

		switch level {
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "info":
			log.SetLevel(log.InfoLevel)
		case "warn":
			log.SetLevel(log.WarnLevel)
		case "error":
			log.SetLevel(log.ErrorLevel)
		default:
			log.SetLevel(log.ErrorLevel) // Default to Info if not recognized
		}
	} else {
		log.SetLevel(log.WarnLevel) // Default log level
	}
	// Set log format
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Set output to stderr
	log.SetOutput(os.Stderr)

	log.Debug("Logger initialized with level: ", log.GetLevel())
}