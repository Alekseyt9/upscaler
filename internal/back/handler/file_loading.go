package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Alekseyt9/upscaler/internal/back/services/s3stor"
)

// POST
func (h *FrontHandler) FinishFilesLoading(w http.ResponseWriter, r *http.Request) {
	// recieve right files ids and names
	var links []s3stor.Link
	if err := json.NewDecoder(r.Body).Decode(&links); err != nil {
		http.Error(w, "deserializing links", http.StatusBadRequest)
		return
	}

	// all files have already loaded to s3

	// create messages for kafka
	// store messages to transaction outbox
	// store tasks to db

}

// GET
func (h *FrontHandler) GetRequisites(w http.ResponseWriter, r *http.Request) {
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

	/*
		urls := make([]string, 0)
		for _, v := range links {
			urls = append(urls, v.Url)
		}*/

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(links); err != nil {
		http.Error(w, "error encoding json", http.StatusInternalServerError)
	}
}
