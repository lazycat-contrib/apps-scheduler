package web

import (
	"embed"
	"io/fs"
	"net/http"

	"apps-scheduler/internal/auth"
	"apps-scheduler/internal/biz"
	"apps-scheduler/internal/handlers"
	"apps-scheduler/internal/version"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

//go:embed public/*
var publicFS embed.FS

type Server struct {
	echo         *echo.Echo
	useCase      *biz.UseCase
	oidcProvider *auth.OIDCProvider
	publicContent fs.FS
}

func NewServer(useCase *biz.UseCase) *Server {
	e := echo.New()
	e.HideBanner = true

	// Try to initialize OIDC provider
	oidcProvider, err := auth.NewOIDCProvider()
	if err != nil {
		log.Warn().Err(err).Msg("OIDC not configured, using header-based auth only")
	}

	publicContent, _ := fs.Sub(publicFS, "public")

	server := &Server{
		echo:          e,
		useCase:       useCase,
		oidcProvider:  oidcProvider,
		publicContent: publicContent,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

func (s *Server) setupMiddleware() {
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.RequestID())
	s.echo.Use(middleware.Logger())

	// Session middleware to restore cookies
	s.echo.Use(auth.SessionMiddleware())
}

func (s *Server) setupRoutes() {
	// Static files
	staticHandler := http.FileServer(http.FS(s.publicContent))
	s.echo.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", staticHandler)))

	// Public routes
	s.echo.GET("/login", s.serveFile("login.html"))
	s.echo.GET("/logout", auth.HandleLogout)

	// OIDC routes
	if s.oidcProvider != nil {
		s.echo.GET(auth.GetOIDCBasePath()+"/login", s.oidcProvider.HandleLogin)
		s.echo.GET(auth.GetOIDCCallbackPath(), s.oidcProvider.HandleCallback)
	}

	// Protected routes
	protected := s.echo.Group("")
	protected.Use(auth.AuthMiddleware(s.oidcProvider))

	// Pages
	protected.GET("/", s.serveFile("index.html"))
	protected.GET("/settings", s.serveFile("settings.html"))

	// API routes
	api := protected.Group("/api")

	// User info
	userInfoHandler := handlers.NewUserInfoHandler()
	api.GET("/userinfo", userInfoHandler.GetUserInfo)

	// Apps
	appHandler := handlers.NewAppHandler()
	api.GET("/apps", appHandler.ListApps)
	api.POST("/apps/:appId/resume", appHandler.ResumeApp)
	api.POST("/apps/:appId/pause", appHandler.PauseApp)

	// Schedules
	scheduleHandler := handlers.NewScheduleHandler(s.useCase)
	api.GET("/schedules", scheduleHandler.ListSchedules)
	api.POST("/schedules", scheduleHandler.CreateSchedule)
	api.PUT("/schedules/:id", scheduleHandler.UpdateSchedule)
	api.DELETE("/schedules/:id", scheduleHandler.DeleteSchedule)
	api.POST("/schedules/:id/toggle", scheduleHandler.ToggleSchedule)

	// Notify
	notifyHandler := handlers.NewNotifyHandler(s.useCase)
	api.GET("/notify/config", notifyHandler.GetConfig)
	api.POST("/notify/config", notifyHandler.SaveConfig)
	api.POST("/notify/test", notifyHandler.TestNotify)

	// Version - public API (no auth required)
	s.echo.GET("/api/version", s.getVersion)
}

func (s *Server) serveFile(filename string) echo.HandlerFunc {
	return func(c echo.Context) error {
		data, err := fs.ReadFile(s.publicContent, filename)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Page not found")
		}
		return c.HTMLBlob(http.StatusOK, data)
	}
}

func (s *Server) getVersion(c echo.Context) error {
	return c.JSON(http.StatusOK, version.Get())
}

func (s *Server) Start(addr string) error {
	log.Info().Str("addr", addr).Msg("Starting HTTP server")
	return s.echo.Start(addr)
}

func (s *Server) Shutdown() error {
	return s.echo.Close()
}
