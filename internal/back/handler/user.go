// Package handler provides HTTP handlers for handling API requests related to user file uploads,
// file processing state retrieval, and generating pre-signed URLs for file uploads.
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Alekseyt9/upscaler/internal/back/handler/middleware/jwtcheker"
	"github.com/Alekseyt9/upscaler/internal/back/model"
)

// CompleteFilesUpload handles the POST request for completing file uploads.
// It extracts the user ID from the JWT token, deserializes the list of uploaded files,
// and creates tasks for processing the files.
func (h *ServerHandler) CompleteFilesUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := jwtcheker.GetUserID(r)
	if !ok {
		h.log.Error("GetUserID", "not found", ok)
		http.Error(w, "GetUserID", http.StatusInternalServerError)
		return
	}
	h.log.Info("GetUserID", "userID", userID)

	var fileInfos []model.UploadedFile
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&fileInfos); err != nil {
		h.log.Error("deserializing fileInfos", "error", err)
		http.Error(w, "deserializing fileInfos", http.StatusInternalServerError)
		return
	}

	err := h.us.CreateTasks(r.Context(), fileInfos, userID)
	if err != nil {
		h.log.Error("us.CreateTasks", "error", err)
		http.Error(w, "us.CreateTasks", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetUploadURLs handles the GET request to generate pre-signed URLs for file uploads.
// It extracts the user ID from the JWT token and generates pre-signed URLs for file upload based on the provided count.
func (h *ServerHandler) GetUploadURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := jwtcheker.GetUserID(r)
	if !ok {
		h.log.Error("GetUserID", "not found", ok)
		http.Error(w, "GetUserID", http.StatusBadRequest)
		return
	}
	h.log.Info("GetUserID", "userID", userID)

	count, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err != nil {
		http.Error(w, "'count' parameter must be a valid integer", http.StatusBadRequest)
		return
	}

	if count > 10 {
		http.Error(w, "max files count = 10", http.StatusBadRequest)
		return
	}

	links, err := h.s3.GetPresigned(count)
	if err != nil {
		http.Error(w, "GetPresigned error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(links); err != nil {
		http.Error(w, "error encoding json", http.StatusInternalServerError)
	}
}

// GetState handles the GET request to retrieve the state of uploaded files for the user.
// It extracts the user ID from the JWT token and retrieves the current state of the user's files from the store.
func (h *ServerHandler) GetState(w http.ResponseWriter, r *http.Request) {
	userID, ok := jwtcheker.GetUserID(r)
	if !ok {
		h.log.Error("GetUserID", "not found", ok)
		http.Error(w, "GetUserID", http.StatusInternalServerError)
		return
	}
	h.log.Info("GetUserID", "userID", userID)

	items, err := h.store.GetState(r.Context(), userID)
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
