package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "time/tzdata" // Embed timezone database for cross-platform compatibility

	"apps-scheduler/internal/biz"
	"apps-scheduler/internal/pkg/zlog"
	"apps-scheduler/internal/version"
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

	// Log version info
	log.Info().
		Str("version", version.Version).
		Str("git_commit", version.GitCommit).
		Str("build_time", version.BuildTime).
		Msg("Starting apps-scheduler")

	// Log timezone information for debugging
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "not set (using system default)"
	}
	now := time.Now()
	loc, _ := time.LoadLocation(tz)
	if loc != nil {
		now = now.In(loc)
	}
	log.Info().
		Str("TZ_env", tz).
		Str("current_time", now.Format(time.RFC3339)).
		Str("timezone", now.Location().String()).
		Msg("Timezone configuration")

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
