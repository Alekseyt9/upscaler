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
}

func New(s3 s3store.S3Store, log *slog.Logger, store store.Store) *FrontHandler {
	return &FrontHandler{
		s3:    s3,
		log:   log,
		store: store,
	}
}
