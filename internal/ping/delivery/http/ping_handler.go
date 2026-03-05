package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type PingHandler struct {
}

func NewPingHandler() *PingHandler {
	return &PingHandler{}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	bodyString := strings.TrimSpace(string(bodyBytes))
	if bodyString != "ping" {
		http.Error(w, "No ping? Plaki-plaki", http.StatusBadRequest)
		return
	}
	response := map[string]string{"message": "pong"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Error in ping-pong", http.StatusInternalServerError)
	}
}
