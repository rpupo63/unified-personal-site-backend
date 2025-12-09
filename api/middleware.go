package api

import (
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/rpupo63/unified-personal-site-backend/errs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type authMiddleware struct {
	responder Responder
}

func newAuthMiddleware() authMiddleware {
	logger := log.With().Str("handlerName", "authMiddleware").Logger()
	return authMiddleware{
		responder: NewResponder(logger),
	}
}

func (m authMiddleware) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			m.responder.WriteError(w, errs.Unauthorized)
			return
		}

		userID := strings.TrimPrefix(authHeader, "Bearer ")
		if userID == "" {
			m.responder.WriteError(w, errs.Unauthorized)
			return
		}

		ctx := r.Context()
		updatedCtx := ctxWithUserID(ctx, userID)
		updatedReq := r.WithContext(updatedCtx)
		next.ServeHTTP(w, updatedReq)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	if !w.wroteHeader {
		w.status = statusCode
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(statusCode)
	}
}
func LogInternalServerErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srw := &statusResponseWriter{ResponseWriter: w, status: 200}

		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Interface("panic", err).
					Str("stack", string(debug.Stack())).
					Msg("Recovered from panic")

				// Write 500 if nothing written yet
				if !srw.wroteHeader {
					srw.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(srw, r)

		// Optionally log 500s that weren't panics (e.g. manually set by handlers)
		if srw.status == http.StatusInternalServerError {
			log.Error().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("500 error response")
		}
	})
}

// CORSCheckMiddleware checks if the request is blocked by CORS and returns a proper error
func CORSCheckMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// If no origin header, it's likely a same-origin request
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check if origin is in allowed list
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			// If not allowed and it's a preflight request, return error
			if !allowed && r.Method == "OPTIONS" {
				responder := NewResponder(log.Logger)
				responder.WriteError(w, errs.NewCORSError(origin))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// corsMiddleware handles CORS headers for allowed origins
func corsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ColoredHTTPLoggingMiddleware logs HTTP requests with colored output based on status codes
func ColoredHTTPLoggingMiddleware(next http.Handler) http.Handler {
	// Set up colored console writer for development
	colorLogger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Logger()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		srw := &statusResponseWriter{ResponseWriter: w, status: 200}

		next.ServeHTTP(srw, r)

		duration := time.Since(start)

		// Color-code based on HTTP status codes
		var logEvent *zerolog.Event
		switch {
		case srw.status >= 500:
			logEvent = colorLogger.Error()
		case srw.status >= 400:
			logEvent = colorLogger.Warn()
		case srw.status >= 300:
			logEvent = colorLogger.Info()
		default:
			logEvent = colorLogger.Info()
		}

		logEvent.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", srw.status).
			Dur("duration", duration).
			Str("remote_addr", r.RemoteAddr).
			Msg("HTTP Request")
	})
}
