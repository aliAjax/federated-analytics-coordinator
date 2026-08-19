package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	aggregate "github.com/example/federated-analytics-coordinator/internal/aggregation/domain"
	federation "github.com/example/federated-analytics-coordinator/internal/federation/domain"
	"github.com/example/federated-analytics-coordinator/internal/platform/config"
	protocol "github.com/example/federated-analytics-coordinator/internal/protocol/domain"
	service "github.com/example/federated-analytics-coordinator/internal/protocol/application"
	"io"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"
)

type envelope struct {
	Data      any       `json:"data,omitempty"`
	Error     *apiError `json:"error,omitempty"`
	RequestID string    `json:"request_id"`
	At        time.Time `json:"at"`
}
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type Server struct {
	cfg     config.Config
	service *service.Service
	mux     *http.ServeMux
}

func New(cfg config.Config, s *service.Service) *Server {
	x := &Server{cfg: cfg, service: s, mux: http.NewServeMux()}
	x.routes()
	return x
}
func (s *Server) Handler() http.Handler { return s.recover(s.mux) }
func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		s.json(w, 200, rid(r), map[string]string{"status": "ok"})
	})
	s.mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		s.json(w, 200, rid(r), map[string]string{"status": "ready"})
	})
	s.mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	s.mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	s.mux.HandleFunc("POST /api/v1/participants", s.registerParticipant)
	s.mux.HandleFunc("POST /api/v1/tasks", s.createTask)
	s.mux.HandleFunc("GET /api/v1/tasks/{id}", s.getTask)
	s.mux.HandleFunc("POST /api/v1/tasks/{id}/start", s.start)
	s.mux.HandleFunc("POST /api/v1/tasks/{id}/shares", s.share)
	s.mux.HandleFunc("POST /api/v1/tasks/{id}/aggregate", s.aggregate)
	s.mux.HandleFunc("POST /api/v1/tasks/{id}/abort", s.abort)
	s.mux.HandleFunc("GET /api/v1/results/{id}", s.result)
	s.mux.HandleFunc("GET /api/v1/audit/export", s.audit)
}
func tenant(r *http.Request) (string, error) {
	v := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	if v == "" {
		return "", fmt.Errorf("X-Tenant-ID is required")
	}
	return v, nil
}
func rid(r *http.Request) string {
	if v := r.Header.Get("X-Request-ID"); v != "" {
		return v
	}
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
func (s *Server) decode(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, s.cfg.MaxBodyBytes)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(dst); e != nil {
		return e
	}
	if e := d.Decode(&struct{}{}); e != io.EOF {
		return fmt.Errorf("one JSON document required")
	}
	return nil
}
func (s *Server) json(w http.ResponseWriter, status int, id string, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", id)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: v, RequestID: id, At: time.Now().UTC()})
}
func (s *Server) fail(w http.ResponseWriter, status int, id, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Error: &apiError{code, msg}, RequestID: id, At: time.Now().UTC()})
}
func (s *Server) registerParticipant(w http.ResponseWriter, r *http.Request) {
	id := rid(r)
	t, e := tenant(r)
	if e != nil {
		s.fail(w, 400, id, "invalid_tenant", e.Error())
		return
	}
	var p federation.Participant
	if e = s.decode(r, &p); e != nil {
		s.fail(w, 400, id, "invalid_request", e.Error())
		return
	}
	v, e := s.service.RegisterParticipant(t, p)
	if e != nil {
		s.fail(w, 422, id, "participant_rejected", e.Error())
		return
	}
	s.json(w, 201, id, v)
}
func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	id := rid(r)
	t, e := tenant(r)
	if e != nil {
		s.fail(w, 400, id, "invalid_tenant", e.Error())
		return
	}
	var c service.CreateTask
	if e = s.decode(r, &c); e != nil {
		s.fail(w, 400, id, "invalid_request", e.Error())
		return
	}
	v, e := s.service.CreateTask(t, c)
	if e != nil {
		s.fail(w, 422, id, "task_rejected", e.Error())
		return
	}
	s.json(w, 201, id, v)
}
func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	id := rid(r)
	t, e := tenant(r)
	if e != nil {
		s.fail(w, 400, id, "invalid_tenant", e.Error())
		return
	}
	v, e := s.service.GetTask(t, r.PathValue("id"))
	if e != nil {
		s.fail(w, 404, id, "not_found", e.Error())
		return
	}
	s.json(w, 200, id, v)
}
func (s *Server) start(w http.ResponseWriter, r *http.Request) {
	id := rid(r)
	t, e := tenant(r)
	if e != nil {
		s.fail(w, 400, id, "invalid_tenant", e.Error())
		return
	}
	v, e := s.service.Start(t, r.PathValue("id"))
	if e != nil {
		if errors.Is(e, protocol.ErrNotFound) {
			s.fail(w, 404, id, "not_found", e.Error())
			return
		}
		s.fail(w, 422, id, "start_rejected", e.Error())
		return
	}
	s.json(w, 200, id, v)
}
func (s *Server) share(w http.ResponseWriter, r *http.Request) {
	id := rid(r)
	t, e := tenant(r)
	if e != nil {
		s.fail(w, 400, id, "invalid_tenant", e.Error())
		return
	}
	var v aggregate.Share
	if e = s.decode(r, &v); e != nil {
		s.fail(w, 400, id, "invalid_request", e.Error())
		return
	}
	v.TaskID = r.PathValue("id")
	saved, e := s.service.UploadShare(t, v)
	if e != nil {
		s.fail(w, 422, id, "share_rejected", e.Error())
		return
	}
	s.json(w, 202, id, saved)
}
func (s *Server) aggregate(w http.ResponseWriter, r *http.Request) {
	id := rid(r)
	t, e := tenant(r)
	if e != nil {
		s.fail(w, 400, id, "invalid_tenant", e.Error())
		return
	}
	v, e := s.service.Aggregate(t, r.PathValue("id"))
	if e != nil {
		s.fail(w, 422, id, "aggregate_rejected", e.Error())
		return
	}
	s.json(w, 200, id, v)
}
func (s *Server) abort(w http.ResponseWriter, r *http.Request) {
	id := rid(r)
	t, e := tenant(r)
	if e != nil {
		s.fail(w, 400, id, "invalid_tenant", e.Error())
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if e = s.decode(r, &req); e != nil {
		s.fail(w, 400, id, "invalid_request", e.Error())
		return
	}
	v, e := s.service.Abort(t, r.PathValue("id"), req.Reason)
	if e != nil {
		s.fail(w, 422, id, "abort_rejected", e.Error())
		return
	}
	s.json(w, 200, id, v)
}
func (s *Server) result(w http.ResponseWriter, r *http.Request) {
	id := rid(r)
	t, e := tenant(r)
	if e != nil {
		s.fail(w, 400, id, "invalid_tenant", e.Error())
		return
	}
	v, e := s.service.Result(t, r.PathValue("id"))
	if e != nil {
		s.fail(w, 404, id, "not_found", e.Error())
		return
	}
	s.json(w, 200, id, v)
}
func (s *Server) audit(w http.ResponseWriter, r *http.Request) {
	id := rid(r)
	t, e := tenant(r)
	if e != nil {
		s.fail(w, 400, id, "invalid_tenant", e.Error())
		return
	}
	s.json(w, 200, id, s.service.Audit(t))
}
func (s *Server) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				s.fail(w, 500, rid(r), "internal_error", fmt.Sprint(v))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
