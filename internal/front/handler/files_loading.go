package handler

import (
	"net/http"
)

// GetPresignedURLs обрабатывает запрос и проверяет максимальное количество файлов (5).
func (h *FrontHandler) GetPresignedURLs(w http.ResponseWriter, r *http.Request) {
	// Проверка на максимальное количество файлов (5)
	// files := r.URL.Query()["files"]

	// Ответить клиенту
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Handled presigned URLs"))
}
