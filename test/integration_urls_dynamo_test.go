package test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/luizpaulino/url-shortener/internal/urls"
	"github.com/luizpaulino/url-shortener/pkg/log"
)

type fixedGen struct{}

func (fixedGen) Generate(ctx context.Context) (string, error) { return "abc123", nil }

func TestIntegration_URLs_Flow(t *testing.T) {
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
	repo := urls.NewDynamoRepo(endpoint, table, logger)
	svc := urls.NewDefaultService(repo, fixedGen{}, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item, err := svc.Create(ctx, "https://example.com/x")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if item.ShortCode != "abc123" {
		t.Fatalf("expected abc123, got %s", item.ShortCode)
	}

	got, err := svc.Get(ctx, "abc123")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.LongURL != "https://example.com/x" || got.Clicks != 0 {
		t.Fatalf("unexpected item: %#v", got)
	}

	long, err := svc.ResolveAndIncrement(ctx, "abc123")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if long != "https://example.com/x" {
		t.Fatalf("unexpected redirect long url: %s", long)
	}

	got2, err := svc.Get(ctx, "abc123")
	if err != nil {
		t.Fatalf("get2: %v", err)
	}
	if got2.Clicks != 1 {
		t.Fatalf("expected clicks=1, got %d", got2.Clicks)
	}
}
