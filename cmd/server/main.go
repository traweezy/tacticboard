package main

import (
	"go.uber.org/fx"

	"github.com/traweezy/tacticboard/internal/app"
)

func main() {
	app := fx.New(app.Module)
	app.Run()
}
