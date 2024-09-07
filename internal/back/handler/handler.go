package handler

import "github.com/Alekseyt9/upscaler/internal/back/services/s3stor"

type FrontHandler struct {
	s3 s3stor.S3Store
}

func New(s3 s3stor.S3Store) *FrontHandler {
	return &FrontHandler{s3: s3}
}
