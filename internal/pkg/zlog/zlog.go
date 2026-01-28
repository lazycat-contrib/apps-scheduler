package zlog

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Init(logDir string) {
	if logDir == "" {
		logDir = "/var/log/apps-scheduler"
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Warn().Err(err).Msg("Failed to create log directory, using console only")
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
			With().Timestamp().Caller().Logger()
		return
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "app.log"),
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	multi := io.MultiWriter(consoleWriter, fileWriter)

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = zerolog.New(multi).With().Timestamp().Caller().Logger()
}
