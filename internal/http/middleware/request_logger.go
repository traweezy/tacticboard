package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	// RequestIDHeader is the canonical header that carries a request identifier.
	RequestIDHeader = "X-Request-ID"
	requestIDKey    = "request-id"
)

// HTTPMetrics encapsulates server-side counters and histograms.
type HTTPMetrics struct {
	requestDuration metric.Float64Histogram
	requestCounter  metric.Int64Counter
}

// NewHTTPMetrics constructs histogram and counter instruments for HTTP traffic.
func NewHTTPMetrics(meter metric.Meter) (*HTTPMetrics, error) {
	duration, err := meter.Float64Histogram(
		"http.server.duration",
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of HTTP requests handled by the server"),
	)
	if err != nil {
		return nil, err
	}

	counter, err := meter.Int64Counter(
		"http.server.requests",
		metric.WithDescription("Count of HTTP requests handled by the server"),
	)
	if err != nil {
		return nil, err
	}

	return &HTTPMetrics{
		requestDuration: duration,
		requestCounter:  counter,
	}, nil
}

// RequestLogger instruments incoming HTTP traffic with tracing, metrics, and structured logging.
func RequestLogger(log *zap.Logger, tracer trace.Tracer, metrics *HTTPMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := c.GetHeader(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Set(requestIDKey, reqID)
		c.Writer.Header().Set(RequestIDHeader, reqID)

		ctx, span := tracer.Start(
			c.Request.Context(),
			fmt.Sprintf("%s %s", c.Request.Method, c.FullPath()),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		attrs := []attribute.KeyValue{
			semconv.HTTPRoute(route),
			semconv.HTTPRequestMethodKey.String(c.Request.Method),
			semconv.URLPath(c.Request.URL.Path),
			attribute.Int("http.status_code", status),
		}

		span.SetAttributes(attrs...)
		if status >= http.StatusInternalServerError {
			span.SetStatus(codes.Error, http.StatusText(status))
		} else {
			span.SetStatus(codes.Ok, http.StatusText(status))
		}
		span.End()

		if metrics != nil {
			metrics.record(ctx, route, c.Request.Method, status, latency)
		}

		traceID := span.SpanContext().TraceID()
		log.Info("http request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("route", route),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("request_id", reqID),
			zap.String("trace_id", traceID.String()),
		)
	}
}

func (m *HTTPMetrics) record(ctx context.Context, route, method string, status int, latency time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("http.route", route),
		attribute.String("http.method", method),
		attribute.Int("http.status_code", status),
	}
	options := metric.WithAttributes(attrs...)
	m.requestCounter.Add(ctx, 1, options)
	m.requestDuration.Record(ctx, float64(latency.Milliseconds()), options)
}
