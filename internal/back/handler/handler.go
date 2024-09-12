package handler

import "github.com/Alekseyt9/upscaler/internal/back/services/s3store"

type FrontHandler struct {
	s3 s3store.S3Store
}

func New(s3 s3store.S3Store) *FrontHandler {
	return &FrontHandler{s3: s3}
}
