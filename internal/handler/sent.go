package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"insider-challenge/pkg/config"
	domain "insider-challenge/pkg/domain"
	apperrors "insider-challenge/pkg/errors"
)

// MessageWithCache extends the Message struct with Redis cache inf
type MessageWithCache struct {
	domain.Message
	CachedSentAt    *time.Time `json:"cached_sent_at,omitempty"`
	CachedMessageID string     `json:"cached_message_id,omitempty"`
}

// PaginatedMessagesResponse represents paginated response of messages with cache inf
type PaginatedMessagesResponse struct {
	Messages []MessageWithCache `json:"messages"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int64              `json:"total"`
}

// @Summary Get sent messages
// @Description Retrieves a paginated list of sent messages
// @Tags message
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Number of items per page (default: 10, max: 100)"
// @Success 200 {object} PaginatedMessagesResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /sent [get]
func (h *Handler) handleSent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response := ErrorResponse{Error: "Method not allowed"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse pagination param
	page := 1
	pageSize := h.cfg.DefaultPageSize

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			if ps > h.cfg.MaxPageSize {
				ps = h.cfg.MaxPageSize
			}
			pageSize = ps
		}
	}

	messages, total, err := h.service.GetSentMessages(page, pageSize)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			switch appErr.Unwrap() {
			case apperrors.ErrMessageNotFound:
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "No messages found"})
			case apperrors.ErrDatabaseOperation:
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Database operation failed"})
			default:
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Error getting sent messages: %v", err)})
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Error getting sent messages: %v", err)})
		return
	}

	// Create a context for redis operations
	ctx := context.Background()

	// Convert messages to include redis cache inf
	messagesWithCache := make([]MessageWithCache, len(messages))
	for i, msg := range messages {
		messagesWithCache[i] = MessageWithCache{
			Message: msg,
		}

		// Get cached information from redis
		if cache, err := config.GetMessageCache(ctx, msg.ID.String()); err == nil {
			sentAt := time.Unix(cache.SentAt, 0)
			messagesWithCache[i].CachedSentAt = &sentAt
			messagesWithCache[i].CachedMessageID = cache.MessageID
		}
	}

	response := PaginatedMessagesResponse{
		Messages: messagesWithCache,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		response := ErrorResponse{Error: fmt.Sprintf("Error encoding response: %v", err)}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
}
