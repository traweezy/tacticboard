package store

import (
	"context"

	"github.com/traweezy/tacticboard/internal/config"
	"github.com/traweezy/tacticboard/internal/model"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Store captures persistence operations required by the service.
type Store interface {
	CreateRoom(ctx context.Context, room model.Room) (model.Room, error)
	GetRoom(ctx context.Context, roomID string) (model.Room, error)
	SaveSnapshot(ctx context.Context, snapshot model.Snapshot) error
	LatestSnapshot(ctx context.Context, roomID string) (model.Snapshot, error)
	AppendOperation(ctx context.Context, op model.Operation) (model.Operation, error)
	OperationsSince(ctx context.Context, roomID string, sinceSeq int64, limit int) ([]model.Operation, error)
}

// Module registers the store implementation.
var Module = fx.Module(
	"store",
	fx.Provide(New),
)

// New returns the configured Store implementation.
func New(cfg config.Config, log *zap.Logger) (Store, error) {
	if cfg.DBEnable {
		store, err := newPostgresStore(cfg.DBDSN)
		if err != nil {
			return nil, err
		}
		log.Info("store initialized", zap.String("driver", "postgres"))
		return store, nil
	}

	log.Info("store initialized", zap.String("driver", "memory"))
	return NewMemoryStore(), nil
}
