package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/luizpaulino/url-shortener/internal/urls"
	"github.com/luizpaulino/url-shortener/pkg/log"
)

type fakeService struct{}

func (f *fakeService) Create(ctx context.Context, longURL string) (urls.URLItem, error) {
	if longURL == "" || longURL == "bad" {
		return urls.URLItem{}, urls.ErrInvalid
	}
	return urls.URLItem{
		ShortCode: "abc123",
		LongURL:   longURL,
		Clicks:    0,
		CreatedAt: time.Unix(0, 0),
	}, nil
}
func (f *fakeService) Get(ctx context.Context, shortCode string) (urls.URLItem, error) {
	if shortCode == "missing" {
		return urls.URLItem{}, urls.ErrNotFound
	}
	return urls.URLItem{ShortCode: shortCode, LongURL: "https://example.com/x", Clicks: 1, CreatedAt: time.Unix(0, 0)}, nil
}
func (f *fakeService) ResolveAndIncrement(ctx context.Context, shortCode string) (string, error) {
	if shortCode == "missing" {
		return "", urls.ErrNotFound
	}
	return "https://example.com/x", nil
}

func TestCreateURL_HTTP_201(t *testing.T) {
	logger := log.New()
	s := NewServerWithDeps(Config{Port: "0"}, logger, &fakeService{})

	body := map[string]string{"longURL": "https://example.com/x"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/urls", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	var got urls.URLItem
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got.ShortCode != "abc123" || got.LongURL != "https://example.com/x" {
		t.Fatalf("unexpected body: %#v", got)
	}
}

func TestGetURL_HTTP_200(t *testing.T) {
	logger := log.New()
	s := NewServerWithDeps(Config{Port: "0"}, logger, &fakeService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/urls/abc123", nil)
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got urls.URLItem
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got.ShortCode != "abc123" {
		t.Fatalf("unexpected shortcode: %s", got.ShortCode)
	}
}

func TestRedirect_HTTP_301(t *testing.T) {
	logger := log.New()
	s := NewServerWithDeps(Config{Port: "0"}, logger, &fakeService{})

	req := httptest.NewRequest(http.MethodGet, "/r/abc123", nil)
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	if loc != "https://example.com/x" {
		t.Fatalf("expected redirect to https://example.com/x, got %s", loc)
	}
}
