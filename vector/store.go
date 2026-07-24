package vector

import (
	"context"
)

type Match struct {
	ID    string
	Score float32
}

// Store defines the unified interface for vector databases.
type Store interface {
	Upsert(ctx context.Context, id string, vec []float32, payload map[string]any) error
	SearchNearest(ctx context.Context, vec []float32, limit int, filter map[string]any) ([]Match, error)
}
