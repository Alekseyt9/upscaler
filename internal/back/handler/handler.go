package handler

import (
	"log/slog"

	"github.com/Alekseyt9/upscaler/internal/back/services/s3store"
	"github.com/Alekseyt9/upscaler/internal/back/services/store"
)

type FrontHandler struct {
	s3    s3store.S3Store
	log   *slog.Logger
	store store.Store
	opt   HandlerOptions
}

type HandlerOptions struct {
	JWTSecret string
}

func New(s3 s3store.S3Store, log *slog.Logger, store store.Store, opt HandlerOptions) *FrontHandler {
	return &FrontHandler{
		s3:    s3,
		log:   log,
		store: store,
		opt:   opt,
	}
}
