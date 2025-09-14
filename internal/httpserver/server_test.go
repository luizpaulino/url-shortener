package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luizpaulino/url-shortener/pkg/log"
)

func TestHealthz_Returns200(t *testing.T) {
	logger := log.New()
	s := NewServer(Config{Port: "0", TableName: "urls"}, logger)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if got := rr.Body.String(); got != "ok" {
		t.Fatalf("expected body 'ok', got %q", got)
	}
}

func TestReadyz_Returns503_WhenDynamoDown(t *testing.T) {
	logger := log.New()
	// endpoint inválido para simular indisponibilidade
	s := NewServer(Config{Port: "0", TableName: "urls", DynamoEndpoint: "http://127.0.0.1:59999"}, logger)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}
