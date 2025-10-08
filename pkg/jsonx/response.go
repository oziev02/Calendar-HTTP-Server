package jsonx

import (
	"encoding/json"
	"net/http"
)

type successEnvelope[T any] struct {
	Result T `json:"result"`
}

type errorEnvelope struct {
	Error string `json:"error"`
}

func WriteOK[T any](w http.ResponseWriter, v T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(successEnvelope[T]{Result: v})
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: msg})
}
