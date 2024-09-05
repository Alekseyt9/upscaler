package run

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Alekseyt9/upscaler/internal/front/config"
)

type PageData struct {
	FrontURL string
}

func Run(cfg *config.Config) error {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	httpRouter := Router(cfg, log)

	log.Info("Server started", "url", cfg.Address)
	err := http.ListenAndServe(cfg.Address, httpRouter)
	if err != nil {
		return err
	}

	return nil
}

func Router(cfg *config.Config, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	setupFileServer(mux, cfg, logger)
	// setupHandlers(mux, s, rm, pm, ws, logger)
	return mux
}

func setupFileServer(mux *http.ServeMux, cfg *config.Config, _ *slog.Logger) {
	contentDir := filepath.Join("..", "..", "internal", "front", "content")

	fs := http.FileServer(http.Dir(contentDir))
	mux.Handle("/content/", http.StripPrefix("/content/", fs))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		tmplPath := filepath.Join(contentDir, "index.html")
		tmpl := template.Must(template.ParseFiles(tmplPath))

		fURL := cfg.Address
		data := PageData{
			FrontURL: fURL,
		}
		w.Header().Set("Content-Type", "text/html")
		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
			return
		}
	})
}
