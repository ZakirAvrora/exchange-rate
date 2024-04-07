package main

import (
	"github.com/ZakirAvrora/exchange-rate/config"
	"github.com/ZakirAvrora/exchange-rate/internal/app"
)

func main() {
	cfg := config.NewConfig(".env")
	app.Run(cfg)
}
