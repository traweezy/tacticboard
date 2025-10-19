package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config captures runtime configuration for the TacticBoard service.
type Config struct {
	AppHost              string   `env:"APP_HOST" envDefault:"0.0.0.0"`
	AppPort              int      `env:"APP_PORT" envDefault:"8080"`
	Environment          string   `env:"APP_ENV" envDefault:"development"`
	JWTSecret            string   `env:"JWT_SECRET,required"`
	ServiceName          string   `env:"SERVICE_NAME" envDefault:"tacticboard"`
	AllowedOrigins       []string `env:"APP_ALLOWED_ORIGINS" envSeparator:","`
	APIRateRPS           float64  `env:"API_RATE_RPS" envDefault:"5"`
	APIRateBurst         int      `env:"API_RATE_BURST" envDefault:"10"`
	ObservabilityEnabled bool     `env:"OBSERVABILITY_ENABLED" envDefault:"true"`
	OTLPEndpoint         string   `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:""`
	OTLPHeaders          []string `env:"OTEL_EXPORTER_OTLP_HEADERS" envSeparator:","`
	OTLPInsecure         bool     `env:"OTEL_EXPORTER_OTLP_INSECURE" envDefault:"false"`
	TraceSamplingRatio   float64  `env:"TRACE_SAMPLING_RATIO" envDefault:"1.0"`
	MetricsIntervalSec   int      `env:"METRICS_EXPORT_INTERVAL_SEC" envDefault:"30"`
	DBEnable             bool     `env:"DB_ENABLE" envDefault:"false"`
	DBDSN                string   `env:"DB_DSN" envDefault:"postgres://postgres:postgres@localhost:5432/tacticboard?sslmode=disable"`
	WSWriteBuffer        int      `env:"WS_WRITE_BUFFER" envDefault:"262144"`
	WSReadLimit          int64    `env:"WS_READ_LIMIT" envDefault:"1048576"`
	SnapshotIntervalSec  int      `env:"SNAPSHOT_INTERVAL_SEC" envDefault:"20"`
	PersistEveryNOps     int      `env:"PERSIST_EVERY_N_OPS" envDefault:"50"`
}

// HTTPAddr returns the host:port combination for binding the HTTP server.
func (c Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.AppHost, c.AppPort)
}

// SnapshotInterval converts the configured seconds into a time.Duration.
func (c Config) SnapshotInterval() time.Duration {
	return time.Duration(c.SnapshotIntervalSec) * time.Second
}

// Load parses environment variables into a Config value enforcing baseline validation.
func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse env: %w", err)
	}

	cfg.Environment = strings.ToLower(strings.TrimSpace(cfg.Environment))
	for i, origin := range cfg.AllowedOrigins {
		cfg.AllowedOrigins[i] = strings.TrimSpace(origin)
	}

	cfg.OTLPEndpoint = strings.TrimSpace(cfg.OTLPEndpoint)
	for i, header := range cfg.OTLPHeaders {
		cfg.OTLPHeaders[i] = strings.TrimSpace(header)
	}

	cfg.ServiceName = strings.TrimSpace(cfg.ServiceName)
	if cfg.ServiceName == "" {
		cfg.ServiceName = "tacticboard"
	}

	if len(cfg.JWTSecret) < 16 {
		return Config{}, fmt.Errorf("jwt secret must be at least 16 characters")
	}

	if cfg.TraceSamplingRatio < 0 || cfg.TraceSamplingRatio > 1 {
		return Config{}, fmt.Errorf("trace sampling ratio must be between 0 and 1")
	}

	if cfg.MetricsIntervalSec <= 0 {
		return Config{}, fmt.Errorf("metrics export interval must be positive")
	}

	if cfg.DBEnable && cfg.DBDSN == "" {
		return Config{}, fmt.Errorf("db enabled but DB_DSN not set")
	}

	if cfg.WSReadLimit <= 0 {
		return Config{}, fmt.Errorf("ws read limit must be positive")
	}

	if cfg.WSWriteBuffer <= 0 {
		return Config{}, fmt.Errorf("ws write buffer must be positive")
	}

	if cfg.PersistEveryNOps <= 0 {
		return Config{}, fmt.Errorf("persist every N ops must be positive")
	}

	if cfg.SnapshotIntervalSec <= 0 {
		return Config{}, fmt.Errorf("snapshot interval must be positive")
	}

	if cfg.APIRateRPS <= 0 {
		return Config{}, fmt.Errorf("api rate rps must be positive")
	}

	if cfg.APIRateBurst <= 0 {
		return Config{}, fmt.Errorf("api rate burst must be positive")
	}

	if cfg.Environment == "production" && len(cfg.AllowedOrigins) == 0 {
		return Config{}, fmt.Errorf("APP_ALLOWED_ORIGINS required in production")
	}

	return cfg, nil
}
