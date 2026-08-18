package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

var appStartTime = time.Now().UTC()

type ComponentStatus string

const (
	StatusUp       ComponentStatus = "up"
	StatusDown     ComponentStatus = "down"
	StatusDegraded ComponentStatus = "degraded"
	StatusDisabled ComponentStatus = "disabled"
)

type ComponentHealth struct {
	Status    ComponentStatus `json:"status"`
	LatencyMs int64           `json:"latency_ms,omitempty"`
	Message   string          `json:"message,omitempty"`
}

type HealthResponse struct {
	Status        string                     `json:"status"`
	UptimeSeconds int64                      `json:"uptime_seconds"`
	Timestamp     time.Time                  `json:"timestamp"`
	Components    map[string]ComponentHealth `json:"components"`
}

type HealthChecker struct {
	checkers map[string]func(ctx context.Context) ComponentHealth
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checkers: make(map[string]func(ctx context.Context) ComponentHealth),
	}
}

func (h *HealthChecker) Register(name string, checkFn func(ctx context.Context) ComponentHealth) {
	h.checkers[name] = checkFn
}

func (h *HealthChecker) Check(ctx context.Context) HealthResponse {
	resp := HealthResponse{
		Status:        "healthy",
		UptimeSeconds: int64(time.Since(appStartTime).Seconds()),
		Timestamp:     time.Now().UTC(),
		Components:    make(map[string]ComponentHealth),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, fn := range h.checkers {
		wg.Add(1)
		go func(compName string, check func(context.Context) ComponentHealth) {
			defer wg.Done()
			checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			start := time.Now()
			result := check(checkCtx)
			if result.LatencyMs == 0 && result.Status != StatusDisabled {
				result.LatencyMs = time.Since(start).Milliseconds()
			}

			mu.Lock()
			resp.Components[compName] = result
			if result.Status == StatusDown {
				if compName == "postgres" {
					resp.Status = "unhealthy"
				} else if resp.Status != "unhealthy" {
					resp.Status = "degraded"
				}
			}
			mu.Unlock()
		}(name, fn)
	}

	wg.Wait()
	return resp
}

// LivenessHandler responde 200 se o processo HTTP do Go estiver ativo.
func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":         "live",
			"uptime_seconds": int64(time.Since(appStartTime).Seconds()),
			"timestamp":      time.Now().UTC(),
		})
	}
}

// ReadinessHandler verifica os subsistemas e retorna 200 (se healthy/degraded) ou 503 (se unhealthy).
func (h *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := h.Check(r.Context())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if health.Status == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_ = json.NewEncoder(w).Encode(health)
	}
}
