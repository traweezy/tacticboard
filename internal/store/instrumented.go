package store

import (
	"context"
	"time"

	"github.com/traweezy/tacticboard/internal/model"
	"github.com/traweezy/tacticboard/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func withInstrumentation(base Store, telemetry *observability.Telemetry, log *zap.Logger) Store {
	tracer := telemetry.TracerProvider.Tracer("github.com/traweezy/tacticboard/store")
	meter := telemetry.MeterProvider.Meter("github.com/traweezy/tacticboard/store")

	duration, err := meter.Float64Histogram(
		"store.operation.duration",
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of store operations"),
	)
	if err != nil {
		log.Warn("store metrics: failed to create duration histogram", zap.Error(err))
	}

	failures, err := meter.Int64Counter(
		"store.operation.failures",
		metric.WithDescription("Count of failed store operations"),
	)
	if err != nil {
		log.Warn("store metrics: failed to create failure counter", zap.Error(err))
	}

	return instrumentedStore{
		Store:  base,
		tracer: tracer,
		metrics: storeMetrics{
			duration: duration,
			failures: failures,
		},
	}
}

type instrumentedStore struct {
	Store
	tracer  trace.Tracer
	metrics storeMetrics
}

type storeMetrics struct {
	duration metric.Float64Histogram
	failures metric.Int64Counter
}

func (s instrumentedStore) CreateRoom(ctx context.Context, room model.Room) (model.Room, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "store.CreateRoom")
	defer span.End()

	result, err := s.Store.CreateRoom(ctx, room)
	s.record(ctx, start, "CreateRoom", span, err)
	return result, err
}

func (s instrumentedStore) GetRoom(ctx context.Context, roomID string) (model.Room, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "store.GetRoom")
	defer span.End()

	result, err := s.Store.GetRoom(ctx, roomID)
	s.record(ctx, start, "GetRoom", span, err)
	return result, err
}

func (s instrumentedStore) SaveSnapshot(ctx context.Context, snapshot model.Snapshot) error {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "store.SaveSnapshot")
	defer span.End()

	err := s.Store.SaveSnapshot(ctx, snapshot)
	s.record(ctx, start, "SaveSnapshot", span, err)
	return err
}

func (s instrumentedStore) LatestSnapshot(ctx context.Context, roomID string) (model.Snapshot, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "store.LatestSnapshot")
	defer span.End()

	result, err := s.Store.LatestSnapshot(ctx, roomID)
	s.record(ctx, start, "LatestSnapshot", span, err)
	return result, err
}

func (s instrumentedStore) AppendOperation(ctx context.Context, op model.Operation) (model.Operation, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "store.AppendOperation")
	defer span.End()

	result, err := s.Store.AppendOperation(ctx, op)
	s.record(ctx, start, "AppendOperation", span, err)
	return result, err
}

func (s instrumentedStore) OperationsSince(ctx context.Context, roomID string, sinceSeq int64, limit int) ([]model.Operation, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "store.OperationsSince")
	defer span.End()

	result, err := s.Store.OperationsSince(ctx, roomID, sinceSeq, limit)
	s.record(ctx, start, "OperationsSince", span, err)
	return result, err
}

func (s instrumentedStore) record(ctx context.Context, start time.Time, operation string, span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	if s.metrics.duration != nil {
		attr := metric.WithAttributes(attribute.String("store.operation", operation))
		s.metrics.duration.Record(ctx, float64(time.Since(start).Milliseconds()), attr)
	}
	if err != nil && s.metrics.failures != nil {
		attr := metric.WithAttributes(attribute.String("store.operation", operation))
		s.metrics.failures.Add(ctx, 1, attr)
	}
}
