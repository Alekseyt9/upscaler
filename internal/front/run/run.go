package run

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/gin-contrib/gzip"

	"github.com/Alekseyt9/upscaler/internal/front/config"
	"github.com/gin-gonic/gin"
)

type PageData struct {
	FrontURL string
}

func Run(cfg *config.Config) error {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	r := Router(cfg, log)

	log.Info("Server started", "url", cfg.Address)
	err := r.Run(cfg.Address)
	if err != nil {
		return err
	}

	return nil
}

func Router(cfg *config.Config, log *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	setupFileServer(r, cfg)
	//setupHandlers(r, s, rm, pm, ws, log)
	return r
}

func setupFileServer(r *gin.Engine, cfg *config.Config) {
	r.Use(gzip.Gzip(gzip.BestCompression))

	contentDir := filepath.Join("..", "..", "internal", "front", "content")
	r.StaticFS("/content", http.Dir(contentDir))

	r.GET("/", func(c *gin.Context) {
		if c.Request.URL.Path != "/" {
			c.String(http.StatusNotFound, "Page not found")
			return
		}
		tmplPath := filepath.Join(contentDir, "index.html")
		tmpl := template.Must(template.ParseFiles(tmplPath))

		fURL := cfg.Address
		data := PageData{
			FrontURL: fURL,
		}
		c.Writer.Header().Set("Content-Type", "text/html")
		err := tmpl.Execute(c.Writer, data)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to render template: %v", err)
			return
		}
	})
}
