package qdrant

import (
	"context"

	"github.com/timescan/timescan/vector"
)

type Config struct {
	Addr       string
	Collection string
}

type Store struct {
	config Config
}

func NewStore(config Config) (*Store, error) {
	return &Store{config: config}, nil
}

func (s *Store) Upsert(ctx context.Context, id string, vec []float32, payload map[string]any) error {
	// In a full implementation, this uses the Qdrant gRPC or HTTP client.
	return nil
}

func (s *Store) SearchNearest(ctx context.Context, vec []float32, limit int, filter map[string]any) ([]vector.Match, error) {
	// Stub search
	return []vector.Match{}, nil
}
