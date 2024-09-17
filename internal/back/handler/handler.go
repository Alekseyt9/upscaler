package handler

import (
	"log/slog"

	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	"github.com/Alekseyt9/upscaler/internal/back/services/userserv"
	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
)

type FrontHandler struct {
	s3    s3store.S3Store
	log   *slog.Logger
	store store.Store
	opt   HandlerOptions
	us    *userserv.UserService
}

type HandlerOptions struct {
	JWTSecret string
}

func New(s3 s3store.S3Store, log *slog.Logger, store store.Store, opt HandlerOptions, us *userserv.UserService) *FrontHandler {
	return &FrontHandler{
		s3:    s3,
		log:   log,
		store: store,
		opt:   opt,
		us:    us,
	}
}
