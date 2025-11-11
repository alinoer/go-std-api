package handlers

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, ErrorResponse{Error: message})
}

func WriteSuccess(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, SuccessResponse{Data: data})
}

func WriteMessage(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusOK, SuccessResponse{Message: message})
}