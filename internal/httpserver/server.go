package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/luizpaulino/url-shortener/internal/urls"
)

type Config struct {
	Port           string
	DynamoEndpoint string
	TableName      string
}

type Server struct {
	Router *chi.Mux
	log    *slog.Logger
	repo   urls.ReadinessRepo // usado só no /readyz
	cfg    Config

	// svc é injetado nos testes e usado pelos endpoints de negócio
	svc urls.Service
}

func NewServer(cfg Config, logger *slog.Logger) *Server {
	r := chi.NewRouter()

	s := &Server{
		Router: r,
		log:    logger,
		cfg:    cfg,
		repo:   urls.NewDynamoReadinessRepo(cfg.DynamoEndpoint, cfg.TableName, logger),
	}

	// Middlewares base
	r.Use(s.requestLogMiddleware())

	// Liveness/Readiness
	r.Get("/healthz", s.handleHealthz)
	r.Get("/readyz", s.handleReadyz)

	// Endpoints de negócio (stubs enquanto o service não estiver setado)
	r.Post("/v1/urls", s.handleCreateURL)
	r.Get("/v1/urls/{shortCode}", s.handleGetURL)
	r.Get("/r/{shortCode}", s.handleRedirect)

	return s
}

// Permite injetar um Service falso nos testes HTTP
func NewServerWithDeps(cfg Config, logger *slog.Logger, svc urls.Service) *Server {
	s := NewServer(cfg, logger)
	s.svc = svc
	return s
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.repo.Ping(ctx); err != nil {
		s.log.Warn("readiness_failed", slog.String("error", err.Error()))
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}

// ------------------- Handlers de URL (stubs) -------------------

type createURLRequest struct {
	LongURL string `json:"longURL"`
}

func (s *Server) handleCreateURL(w http.ResponseWriter, r *http.Request) {
	if s.svc == nil {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}
	var req createURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LongURL == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	item, err := s.svc.Create(r.Context(), req.LongURL)
	if err != nil {
		switch err {
		case urls.ErrInvalid:
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		case urls.ErrConflict:
			http.Error(w, "conflict", http.StatusConflict)
			return
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}

func (s *Server) handleGetURL(w http.ResponseWriter, r *http.Request) {
	if s.svc == nil {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}
	code := chi.URLParam(r, "shortCode")
	if code == "" {
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}
	item, err := s.svc.Get(r.Context(), code)
	if err != nil {
		if err == urls.ErrNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	if s.svc == nil {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}
	code := chi.URLParam(r, "shortCode")
	if code == "" {
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}
	long, err := s.svc.ResolveAndIncrement(r.Context(), code)
	if err != nil {
		if err == urls.ErrNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, long, http.StatusMovedPermanently)
}

// ------------------- Middleware de log -------------------

func (s *Server) requestLogMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(ww, r)
			s.log.Info("http_request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.status),
				slog.Int64("content_length", r.ContentLength),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}
