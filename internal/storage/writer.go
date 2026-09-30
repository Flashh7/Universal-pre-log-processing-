package storage

import (
	"context"

	"github.com/sih/log-platform/internal/normalization"
)

// EventWriter is the interface for persisting validated OCSF events.
type EventWriter interface {
	Write(ctx context.Context, event normalization.NormalizedEvent) error
	Ping(ctx context.Context) error
	Close() error
}
