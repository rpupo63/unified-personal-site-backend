package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rpupo63/unified-personal-site-backend/config"
	"github.com/rpupo63/unified-personal-site-backend/database"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	*http.Server
	startupTime time.Time
}

func NewServer(database database.Database) (Server, error) {
	c := config.New()

	// Get port from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	address := "0.0.0.0:" + port // Bind to 0.0.0.0 for external access

	// Capture startup time
	startupTime := time.Now()

	router := newRouter(database, withConfig(c), withStartupTime(startupTime))

	// Hardcoded timeout values
	readTimeout := 180 * time.Second
	writeTimeout := 180 * time.Second
	idleTimeout := 180 * time.Second

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

	// Apply CORS middleware (must be before routes)
	acceptedOrigins := strings.Split(os.Getenv("ACCEPTED_ORIGINS"), ",")
	chiRouter.Use(CORSCheckMiddleware(acceptedOrigins))
	chiRouter.Use(corsMiddleware(acceptedOrigins))

	// Root endpoint - Message of the Day
	chiRouter.Get("/", rootHandler())

	// Healthcheck endpoint - accessible from any origin
	chiRouter.Get("/healthcheck", healthcheckHandler(router.startupTime))

	// Get backend password from config
	backendPassword := config.GetString(router.config, "BACKEND_PASSWORD", "")

	// Initialize all handlers
	handlers := initializeHandlers(database, backendPassword)

	// Initialize auth middleware
	authMiddleware := newAuthMiddleware()

	// Swagger documentation route
	// Get port from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	swaggerURL := "http://localhost:" + port + "/swagger/doc.json"
	chiRouter.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(swaggerURL), // The url pointing to API definition
	))

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

// rootHandler returns a handler function for the root endpoint
// It displays a "Message of the Day" that adapts based on the client:
// - Plain text for curl/command-line tools
// - HTML terminal-style interface for browsers
func rootHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. A subtle nod in the headers (Visible in DevTools or curl -v)
		w.Header().Set("X-Powered-By", "Arch Linux (btw)")
		w.Header().Set("X-Quantum-State", "Superposition")

		// 2. Personal Easter Eggs
		quotes := []string{
			"SYSTEM STATUS: ONLINE. \nWARNING: Cat detected in server room.",
			"EXECUTION STRATEGY: One way out.",
			"QUANTUM COHERENCE: 99.9%. \nWavefunction has not yet collapsed.",
			"HYPRLAND CONFIG: Loaded. \nTile Layout: Dwindle.",
			"TARGET: The Witness. \nPuzzle status: Unsolved.",
		}

		// Pick a random quote
		rand.Seed(time.Now().UnixNano())
		selectedQuote := quotes[rand.Intn(len(quotes))]

		// 3. Check User-Agent. If it's a browser, render a mini terminal.
		userAgent := r.Header.Get("User-Agent")
		if !strings.Contains(userAgent, "curl") {
			w.Header().Set("Content-Type", "text/html")
			html := fmt.Sprintf(`
        <html>
        <body style="background:#1e1e2e; color:#cdd6f4; font-family: monospace; display:flex; align-items:center; justify-content:center; height:100vh;">
            <div style="border: 1px solid #fab387; padding: 20px; border-radius: 5px;">
                <p style="color:#fab387;">root@api.pupo.codes:~# ./status</p>
                <p>%s</p>
                <span style="animation: blink 1s infinite;">_</span>
            </div>
            <style>@keyframes blink{50%%{opacity:0;}}</style>
        </body>
        </html>`, selectedQuote)
			w.Write([]byte(html))
			return
		}

		// 4. Default plain text response for curl
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(selectedQuote + "\n"))
	}
}

// healthcheckHandler returns a handler function for the healthcheck endpoint
// It returns the current date/time and the server startup time (when this version was deployed)
func healthcheckHandler(startupTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers to allow all origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Create response with current time and startup time
		response := map[string]interface{}{
			"current_time":   time.Now().Format(time.RFC3339),
			"startup_time":   startupTime.Format(time.RFC3339),
			"uptime_seconds": int(time.Since(startupTime).Seconds()),
		}

		// Write JSON response
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error().Err(err).Msg("Error encoding healthcheck response")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
