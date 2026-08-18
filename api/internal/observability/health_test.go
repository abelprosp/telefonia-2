package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthChecker_AllUp(t *testing.T) {
	checker := NewHealthChecker()
	checker.Register("postgres", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Status: StatusUp, LatencyMs: 5}
	})
	checker.Register("rabbitmq", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Status: StatusUp, LatencyMs: 2}
	})

	ctx := context.Background()
	res := checker.Check(ctx)

	if res.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", res.Status)
	}
	if len(res.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(res.Components))
	}
	if res.Components["postgres"].Status != StatusUp {
		t.Errorf("expected postgres status 'up', got '%s'", res.Components["postgres"].Status)
	}
}

func TestHealthChecker_PostgresDown_Unhealthy(t *testing.T) {
	checker := NewHealthChecker()
	checker.Register("postgres", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Status: StatusDown, Message: "connection refused"}
	})

	ctx := context.Background()
	res := checker.Check(ctx)

	if res.Status != "unhealthy" {
		t.Errorf("expected status 'unhealthy' when postgres is down, got '%s'", res.Status)
	}
}

func TestHealthChecker_RabbitMQDown_Degraded(t *testing.T) {
	checker := NewHealthChecker()
	checker.Register("postgres", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Status: StatusUp, LatencyMs: 3}
	})
	checker.Register("rabbitmq", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Status: StatusDown, Message: "broker offline"}
	})

	ctx := context.Background()
	res := checker.Check(ctx)

	if res.Status != "degraded" {
		t.Errorf("expected status 'degraded' when non-critical component is down, got '%s'", res.Status)
	}
}

func TestLivenessHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	handler := LivenessHandler()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "live" {
		t.Errorf("expected body status 'live', got '%v'", body["status"])
	}
}

func TestCorrelationMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrID := CorrelationIDFromContext(r.Context())
		if corrID != "test-req-123" {
			t.Errorf("expected correlation id 'test-req-123', got '%s'", corrID)
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "test-req-123")
	rec := httptest.NewRecorder()

	middleware := CorrelationMiddleware(nextHandler)
	middleware.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") != "test-req-123" {
		t.Errorf("expected response header X-Request-ID to be 'test-req-123', got '%s'", rec.Header().Get("X-Request-ID"))
	}
	if rec.Header().Get("X-Correlation-ID") != "test-req-123" {
		t.Errorf("expected response header X-Correlation-ID to be 'test-req-123', got '%s'", rec.Header().Get("X-Correlation-ID"))
	}
}
