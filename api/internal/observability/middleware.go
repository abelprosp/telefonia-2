package observability

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type correlationCtxKey struct{}

func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationCtxKey{}, id)
}

func CorrelationIDFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(correlationCtxKey{}).(string); ok {
		return val
	}
	return ""
}

// CorrelationMiddleware propaga o Request ID / Correlation ID nos headers e no contexto.
func CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrID := strings.TrimSpace(r.Header.Get("X-Correlation-ID"))
		if corrID == "" {
			corrID = strings.TrimSpace(r.Header.Get("X-Request-ID"))
		}
		if corrID == "" {
			corrID = uuid.New().String()
		}

		ctx := WithCorrelationID(r.Context(), corrID)
		w.Header().Set("X-Correlation-ID", corrID)
		w.Header().Set("X-Request-ID", corrID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// StructuredLoggerMiddleware loga requisições estruturadas via slog com tempo de resposta e correlation id.
func StructuredLoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Ignora logs de polling de liveness para não poluir
			if r.URL.Path == "/health/live" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			corrID := CorrelationIDFromContext(r.Context())
			op := "http.query"
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				op = "http.mutation"
			}
			Observe(op, duration, wrapped.statusCode < 500)
			if strings.Contains(r.URL.Path, "/provider-invoices") {
				Observe("import.api", duration, wrapped.statusCode < 500)
			}
			if strings.Contains(r.URL.Path, "/pipeline") {
				Observe("processing.pipeline.api", duration, wrapped.statusCode < 500)
			}
			if strings.Contains(r.URL.Path, "/tickets") {
				Observe("tickets.api", duration, wrapped.statusCode < 500)
			}
			if r.URL.Path == "/health" || r.URL.Path == "/health/live" {
				Observe("health.api", duration, wrapped.statusCode < 500)
			}

			logLevel := slog.LevelInfo
			if wrapped.statusCode >= 500 {
				logLevel = slog.LevelError
			} else if wrapped.statusCode >= 400 {
				logLevel = slog.LevelWarn
			}

			logger.Log(r.Context(), logLevel, "http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"correlation_id", corrID,
				"client_ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}
