package app

import (
	"github.com/traweezy/tacticboard/internal/config"
	"github.com/traweezy/tacticboard/internal/http"
	"github.com/traweezy/tacticboard/internal/logger"
	"github.com/traweezy/tacticboard/internal/observability"
	"github.com/traweezy/tacticboard/internal/store"
	"github.com/traweezy/tacticboard/internal/util"
	"github.com/traweezy/tacticboard/internal/ws"
	"go.uber.org/fx"
)

// Module composes the application dependency graph.
var Module = fx.Module(
	"tacticboard",
	fx.Provide(
		config.Load,
		ws.NewHub,
	),
	logger.Module,
	observability.Module,
	util.Module,
	store.Module,
	http.Module,
)
