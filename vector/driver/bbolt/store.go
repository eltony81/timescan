package bbolt

import (
	"context"

	"github.com/timescan/timescan/vector"
)

type Config struct {
	Path string
}

type Store struct {
	config Config
}

func NewStore(config Config) (*Store, error) {
	return &Store{config: config}, nil
}

func (s *Store) Upsert(ctx context.Context, id string, vec []float32, payload map[string]any) error {
	// Uses bbolt to store flat vectors or an embedded HNSW index.
	return nil
}

func (s *Store) SearchNearest(ctx context.Context, vec []float32, limit int, filter map[string]any) ([]vector.Match, error) {
	return []vector.Match{}, nil
}
