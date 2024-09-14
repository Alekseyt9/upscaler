package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	//h.log.Info("CompleFilesUpload", "links", links)

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(links); err != nil {
		http.Error(w, "error encoding json", http.StatusInternalServerError)
	}
}

// GET
func (h *FrontHandler) GetState(w http.ResponseWriter, r *http.Request) {

}
