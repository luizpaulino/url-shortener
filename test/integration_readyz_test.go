package test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/luizpaulino/url-shortener/internal/httpserver"
	"github.com/luizpaulino/url-shortener/pkg/log"
)

func TestIntegration_Readyz_OK_WithDynamoLocal(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") == "" {
		t.Skip("RUN_INTEGRATION not set")
	}
	endpoint := os.Getenv("DYNAMO_ENDPOINT")
	if endpoint == "" {
		t.Skip("DYNAMO_ENDPOINT not set")
	}
	table := os.Getenv("URLS_TABLE")
	if table == "" {
		table = "urls"
	}

	logger := log.New()
	s := httpserver.NewServer(httpserver.Config{
		Port:           "0",
		DynamoEndpoint: endpoint,
		TableName:      table,
	}, logger)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%q", rr.Code, rr.Body.String())
	}
}
