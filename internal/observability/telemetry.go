package observability

import (
	"context"
	"strings"
	"time"

	"github.com/traweezy/tacticboard/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otlpmetrichttp "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otlptracehttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

// Telemetry aggregates OpenTelemetry providers used across the service.
type Telemetry struct {
	Enabled        bool
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
}

// Module wires telemetry providers for Fx consumers.
var Module = fx.Module(
	"observability",
	fx.Provide(New),
)

// New configures OpenTelemetry exporters, tracer and meter providers.
func New(lc fx.Lifecycle, cfg config.Config, log *zap.Logger) (*Telemetry, error) {
	if !cfg.ObservabilityEnabled {
		tp := trace.NewNoopTracerProvider()
		mp := noop.NewMeterProvider()
		otel.SetTracerProvider(tp)
		otel.SetMeterProvider(mp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		log.Info("observability disabled; using noop telemetry providers")
		return &Telemetry{
			Enabled:        false,
			TracerProvider: tp,
			MeterProvider:  mp,
		}, nil
	}

	ctx := context.Background()

	res, err := resource.New(
		ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
			attribute.String("service.instance.id", cfg.AppHost),
		),
	)
	if err != nil {
		return nil, err
	}

	traceExporter, err := otlptracehttp.New(ctx, buildTraceExporterOptions(cfg)...)
	if err != nil {
		return nil, err
	}

	metricExporter, err := otlpmetrichttp.New(ctx, buildMetricExporterOptions(cfg)...)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.TraceSamplingRatio))),
	)

	reader := sdkmetric.NewPeriodicReader(
		metricExporter,
		sdkmetric.WithInterval(time.Duration(cfg.MetricsIntervalSec)*time.Second),
	)
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			var shutdownErr error
			if err := mp.Shutdown(ctx); err != nil {
				shutdownErr = multierr.Append(shutdownErr, err)
			}
			if err := tp.Shutdown(ctx); err != nil {
				shutdownErr = multierr.Append(shutdownErr, err)
			}
			return shutdownErr
		},
	})

	log.Info("observability enabled", zap.String("endpoint", cfg.OTLPEndpoint))

	return &Telemetry{
		Enabled:        true,
		TracerProvider: tp,
		MeterProvider:  mp,
	}, nil
}

func buildTraceExporterOptions(cfg config.Config) []otlptracehttp.Option {
	opts := []otlptracehttp.Option{}
	if cfg.OTLPEndpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(cfg.OTLPEndpoint))
	}
	if cfg.OTLPInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if headers := parseHeaders(cfg.OTLPHeaders); len(headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}
	return opts
}

func buildMetricExporterOptions(cfg config.Config) []otlpmetrichttp.Option {
	opts := []otlpmetrichttp.Option{}
	if cfg.OTLPEndpoint != "" {
		opts = append(opts, otlpmetrichttp.WithEndpoint(cfg.OTLPEndpoint))
	}
	if cfg.OTLPInsecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	if headers := parseHeaders(cfg.OTLPHeaders); len(headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(headers))
	}
	return opts
}

func parseHeaders(raw []string) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	headers := make(map[string]string)
	for _, item := range raw {
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" || value == "" {
			continue
		}
		headers[key] = value
	}
	return headers
}
