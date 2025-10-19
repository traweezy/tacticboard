package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/traweezy/tacticboard/internal/config"
	"github.com/traweezy/tacticboard/internal/http/handlers"
	"github.com/traweezy/tacticboard/internal/http/middleware"
)

// Module wires HTTP handlers, router, and server lifecycle.
var Module = fx.Module(
	"http",
	fx.Provide(
		handlers.NewHealthHandler,
		handlers.NewRoomHandler,
		handlers.NewWSHandler,
		NewEngine,
		NewServer,
	),
	fx.Invoke(registerServer),
)

// NewEngine configures the Gin engine with registered routes.
func NewEngine(cfg config.Config, rooms *handlers.RoomHandler, health *handlers.HealthHandler, ws *handlers.WSHandler) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery(), middleware.CORSMiddleware(cfg.AllowedOrigins))

	rateLimiter := middleware.NewIPRateLimiter(cfg.APIRateRPS, cfg.APIRateBurst)

	api := engine.Group("/api")
	api.Use(rateLimiter.Middleware())
	{
		api.GET("/health", health.Handle)
		api.POST("/rooms", rooms.CreateRoom)
		api.GET("/rooms/:id", rooms.GetRoom)
		api.POST("/rooms/:id/share", rooms.ShareRoom)
	}

	engine.GET("/ws/room/:id", ws.Serve)
	return engine
}

// NewServer constructs an *http.Server with sensible timeouts.
func NewServer(cfg config.Config, engine *gin.Engine) *http.Server {
	return &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func registerServer(lc fx.Lifecycle, log *zap.Logger, srv *http.Server) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Info("http server starting", zap.String("addr", srv.Addr))
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatal("http server failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("http server shutting down")
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			return srv.Shutdown(ctx)
		},
	})
}
