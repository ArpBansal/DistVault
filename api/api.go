package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// StorageInterface defines the methods the API needs from storage
type StorageInterface interface {
	StoreData(key string, r io.Reader) error
	Get(key string) (io.ReadCloser, error)
	Delete(id string, key string) error
	GetID() string
}

// APIServer represents the API interface for the distributed storage system
type APIServer struct {
	storage StorageInterface
	address string
	mux     *http.ServeMux
}

// NewAPIServer creates a new API server
func NewAPIServer(storage StorageInterface, address string) *APIServer {
	return &APIServer{
		storage: storage,
		address: address,
		mux:     http.NewServeMux(),
	}
}

// Response formats for the API
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Key     string `json:"key,omitempty"`
}

// Start initializes and starts the API server
func (a *APIServer) Start() error {
	a.mux.HandleFunc("/upload", a.handleUpload)
	a.mux.HandleFunc("/get/", a.handleGet)
	a.mux.HandleFunc("/delete/", a.handleDelete)
	a.mux.HandleFunc("/health", a.handleHealth)

	log.Printf("Starting API server on %s", a.address)
	return http.ListenAndServe(a.address, a.mux)
}

func (a *APIServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)

	key := r.FormValue("key")
	if key == "" {
		key = fmt.Sprintf("file_%d", time.Now().UnixNano())
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to get file from request: "+err.Error())
		return
	}
	defer file.Close()

	err = a.storage.StoreData(key, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to store file: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "File uploaded successfully",
		Key:     key,
	})
}

func (a *APIServer) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Only GET method is allowed")
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/get/")
	if key == "" {
		respondWithError(w, http.StatusBadRequest, "No key provided")
		return
	}

	reader, err := a.storage.Get(key)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "File not found: "+err.Error())
		return
	}
	defer reader.(io.Closer).Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", key))
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = io.Copy(w, reader)
	if err != nil {
		log.Printf("Error streaming file to client: %v", err)
	}
}

func (a *APIServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithError(w, http.StatusMethodNotAllowed, "Only DELETE method is allowed")
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/delete/")
	if key == "" {
		respondWithError(w, http.StatusBadRequest, "No key provided")
		return
	}

	err := a.storage.Delete(a.storage.GetID(), key)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete file: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "File deleted successfully",
		Key:     key,
	})
}

// Handler for health check
func (a *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "API server is running",
	})
}

// Helper function to send JSON responses
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// Helper function to send error responses
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, Response{
		Success: false,
		Message: message,
	})
}
