package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// @Summary Stop message sender
// @Description Stops the automatic message sending process
// @Tags message
// @Accept json
// @Produce json
// @Success 200 {object} StatusResponse
// @Failure 500 {object} ErrorResponse
// @Router /stop [post]
func (h *Handler) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response := ErrorResponse{Error: "Method not allowed"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	if !h.service.IsRunning() {
		response := StatusResponse{Status: "Message sender is not running"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	h.service.StopMessageSender()

	response := StatusResponse{Status: "Message sender stopped"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		response := ErrorResponse{Error: fmt.Sprintf("Error encoding response: %v", err)}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
}
