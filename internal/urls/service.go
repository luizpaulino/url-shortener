package urls

import (
	"context"
	"log/slog"
)

// Service is the business boundary for URL operations.
type Service interface {
	// Create generates a short code for the given long URL and persists it.
	// Should fail with ErrConflict if the short code already exists.
	Create(ctx context.Context, longURL string) (URLItem, error)

	// Get fetches a URL record by its short code.
	Get(ctx context.Context, shortCode string) (URLItem, error)

	// ResolveAndIncrement returns the long URL and increments the click counter atomically.
	ResolveAndIncrement(ctx context.Context, shortCode string) (string, error)
}

// ShortCodeGenerator allows deterministic codes in tests.
type ShortCodeGenerator interface {
	Generate(ctx context.Context) (string, error)
}

// Repo abstracts the persistence.
type Repo interface {
	CreateURL(ctx context.Context, item URLItem) error // must NOT overwrite existing; return ErrConflict
	GetURL(ctx context.Context, shortCode string) (URLItem, error)
	IncrementClicks(ctx context.Context, shortCode string) (int64, error)
}

// DefaultService is a stub. Implement it to make tests pass.
type DefaultService struct {
	repo Repo
	gen  ShortCodeGenerator
	log  *slog.Logger
}

func NewDefaultService(repo Repo, gen ShortCodeGenerator, logger *slog.Logger) *DefaultService {
	return &DefaultService{repo: repo, gen: gen, log: logger}
}

func (s *DefaultService) Create(ctx context.Context, longURL string) (URLItem, error) {
	return URLItem{}, ErrNotImplemented
}

func (s *DefaultService) Get(ctx context.Context, shortCode string) (URLItem, error) {
	return URLItem{}, ErrNotImplemented
}

func (s *DefaultService) ResolveAndIncrement(ctx context.Context, shortCode string) (string, error) {
	return "", ErrNotImplemented
}
