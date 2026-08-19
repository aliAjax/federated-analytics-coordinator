package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/federated-analytics-coordinator/internal/platform/config"
	"github.com/example/federated-analytics-coordinator/internal/platform/store"
	service "github.com/example/federated-analytics-coordinator/internal/protocol/application"
)

func TestStartMissingTaskReturnsNotFound(t *testing.T) {
	cfg := config.Config{Address: ":0", ShutdownTimeout: 1e9, MaxBodyBytes: 1 << 20, MinParticipants: 3, Environment: "test"}
	ids := func() string { return "id" }
	srv := New(cfg, service.New(store.New(), ids))
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/v1/tasks/missing/start", nil)
	req.Header.Set("X-Tenant-ID", "tenant")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing task, got %d", rec.Code)
	}
}
