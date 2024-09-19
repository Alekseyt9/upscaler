// Package main is the entry point for the application.
// It handles the initialization of logging, configuration loading, and server startup.
package main

import (
	"log/slog"
	"os"

	"github.com/Alekseyt9/upscaler/internal/back/config"
	"github.com/Alekseyt9/upscaler/internal/back/run"
)

// main initializes logging, loads the configuration, and starts the server.
// It logs any errors that occur during configuration loading or server startup.
func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("config load error", "error", err)
	}

	err = run.Run(cfg, log)
	if err != nil {
		log.Error("server startup error", "error", err)
	}
}
