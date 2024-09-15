package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Alekseyt9/upscaler/internal/back/handler/middleware/jwtcheker"
	"github.com/Alekseyt9/upscaler/internal/back/model"
	s3stor "github.com/Alekseyt9/upscaler/internal/back/services/s3store"
)

// POST
func (h *FrontHandler) CompleFilesUpload(w http.ResponseWriter, r *http.Request) {
	var links []s3stor.Link
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&links); err != nil {
		h.log.Error("deserializing links", "error", err)
		http.Error(w, "deserializing links", http.StatusInternalServerError)
		return
	}

	err := h.store.CreateTask(model.StoreTask{})
	if err != nil {
		h.log.Error("store.CreateTask", "error", err)
		http.Error(w, "store.CreateTask", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GET
func (h *FrontHandler) GetUploadURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := jwtcheker.GetUserID(r)
	if !ok {
		h.log.Error("GetUserID", "not found", ok)
		http.Error(w, "GetUserID", http.StatusInternalServerError)
	}
	h.log.Info("GetUserID", "userID", userID)

	count, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err != nil {
		http.Error(w, "'count' parameter must be a valid integer", http.StatusBadRequest)
		return
	}

	if count > 10 {
		http.Error(w, "max files count = 10", http.StatusBadRequest)
	}

	links, err := h.s3.GetPresigned(count)
	if err != nil {
		http.Error(w, "GetPresigned error", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(links); err != nil {
		http.Error(w, "error encoding json", http.StatusInternalServerError)
	}
}

// GET
func (h *FrontHandler) GetState(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.GetState(0) // TODO user
	if err != nil {
		h.log.Error("store.GetState", "error", err)
		http.Error(w, "store.GetState", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, "error encoding json", http.StatusInternalServerError)
	}
}
