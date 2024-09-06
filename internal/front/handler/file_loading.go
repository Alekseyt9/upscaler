package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (h *FrontHandler) GetPresignedURLs(w http.ResponseWriter, r *http.Request) {
	count, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err != nil {
		http.Error(w, "'count' parameter must be a valid integer", http.StatusBadRequest)
		return
	}

	if count > 10 {
		http.Error(w, "max count = 10", http.StatusBadRequest)
	}

	links, err := h.s3.GetPresigned(count)
	if err != nil {
		http.Error(w, "GetPresigned error", http.StatusInternalServerError)
	}

	urls := make([]string, 0)
	for _, v := range links {
		urls = append(urls, v.Url)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(urls); err != nil {
		http.Error(w, "error encoding json", http.StatusInternalServerError)
	}
}
