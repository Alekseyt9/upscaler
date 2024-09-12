package run

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Alekseyt9/upscaler/internal/back/config"
)

func Run(cfg *config.Config) error {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	httpRouter := Router(cfg, log)

	log.Info("Server started", "url", cfg.BackAddress)
	err := http.ListenAndServe(cfg.BackAddress, httpRouter)
	if err != nil {
		return err
	}

	return nil
}

func Router(cfg *config.Config, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	setupFileServer(mux, logger)
	// setupHandlers(mux, s, rm, pm, ws, logger)
	return mux
}

func setupFileServer(mux *http.ServeMux, log *slog.Logger) {
	contentDir := filepath.Join("..", "..", "internal", "back", "content")
	log.Info("Serving files from", "contentDir", contentDir)

	fs := http.FileServer(http.Dir(contentDir))
	mux.Handle("/content/", http.StripPrefix("/content/", fs))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("..", "..", "internal", "back", "content", "index.html"))
	})
}
