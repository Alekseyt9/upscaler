package main

import (
	"log/slog"
	"os"

	"github.com/Alekseyt9/upscaler/internal/proc/config"
	"github.com/Alekseyt9/upscaler/internal/proc/run"
)

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
