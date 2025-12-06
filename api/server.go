package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ProNexus-Startup/ProNexus/backend/config"
	"github.com/ProNexus-Startup/ProNexus/backend/database"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type Server struct {
	*http.Server
	startupTime time.Time
}

func NewServer(database database.Database) (Server, error) {
	c := config.New()

	// Ensure correct port is set
	port := config.GetString(c, "PORT", "8080")
	address := fmt.Sprintf("0.0.0.0:%s", port) // Bind to 0.0.0.0 for external access

	// Capture startup time
	startupTime := time.Now()

	router := newRouter(database, withConfig(c), withStartupTime(startupTime))

	// Get timeout values from config with sensible defaults
	readTimeout := time.Duration(config.GetInt(c, "READ_TIMEOUT_SECONDS", 180)) * time.Second
	writeTimeout := time.Duration(config.GetInt(c, "WRITE_TIMEOUT_SECONDS", 180)) * time.Second
	idleTimeout := time.Duration(config.GetInt(c, "IDLE_TIMEOUT_SECONDS", 180)) * time.Second

	server := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  readTimeout,  // Timeout for reading the entire request
		WriteTimeout: writeTimeout, // Timeout for writing the response
		IdleTimeout:  idleTimeout,  // Timeout for idle connections
	}

	return Server{server, startupTime}, nil
}

type router struct {
	config      map[string]string
	startupTime time.Time
}

func withConfig(c map[string]string) func(*router) {
	return func(r *router) {
		r.config = c
	}
}

func withStartupTime(startupTime time.Time) func(*router) {
	return func(r *router) {
		r.startupTime = startupTime
	}
}

func newRouter(database database.Database, opts ...func(*router)) *chi.Mux {
	var router router
	for _, opt := range opts {
		opt(&router)
	}

	chiRouter := chi.NewRouter()
	chiRouter.Use(LogInternalServerErrors)

	// Get backend password from config
	backendPassword := config.GetString(router.config, "BACKEND_PASSWORD", "")

	// Initialize all handlers
	handlers := initializeHandlers(database, backendPassword)

	// Initialize auth middleware
	authMiddleware := newAuthMiddleware()

	// Apply CORS middleware
	acceptedOrigins := strings.Split(os.Getenv("ACCEPTED_ORIGINS"), ",")
	chiRouter.Use(CORSCheckMiddleware(acceptedOrigins))
	chiRouter.Use(corsMiddleware(acceptedOrigins))

	// Setup all route types
	setupFrontendRoutes(chiRouter, handlers, authMiddleware)

	return chiRouter
}

func (s Server) Start(errChannel chan<- error) {
	log.Info().Msgf("Server started on: %s", s.Addr)
	errChannel <- s.ListenAndServe()
}

func (s Server) ShutdownGracefully(timeout time.Duration) {
	log.Info().Msg("Gracefully shutting down...")

	gracefullCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.Shutdown(gracefullCtx); err != nil {
		log.Error().Msgf("Error shutting down the server: %v", err)
	} else {
		log.Info().Msg("HttpServer gracefully shut down")
	}
}
