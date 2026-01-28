package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"apps-scheduler/internal/biz"
	"apps-scheduler/internal/pkg/zlog"
	"apps-scheduler/internal/web"

	"github.com/rs/zerolog/log"
)

const (
	defaultDBPath = "/lzcapp/var/data/apps-scheduler.db"
	defaultLogDir = "/lzcapp/var/data/logs"
)

func main() {
	// Initialize logging
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = defaultLogDir
	}
	os.MkdirAll(logDir, 0755)
	zlog.Init(logDir)

	// Database path
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	// Initialize use case
	useCase, err := biz.NewUseCase(dbPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer useCase.Close()

	// Initialize scheduler
	scheduler := biz.NewScheduler(useCase)
	scheduler.Start()
	defer scheduler.Stop()

	// Initialize web server
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	server := web.NewServer(useCase)

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Info().Msg("Shutting down...")
		scheduler.Stop()
		server.Shutdown()
	}()

	// Start server
	if err := server.Start(addr); err != nil {
		log.Info().Msg("Server stopped")
	}
}
