package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, status int, body Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func Success(w http.ResponseWriter, status int, message string, data any) {
	JSON(w, status, Response{Success: true, Message: message, Data: data})
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, Response{Success: false, Message: message})
}
