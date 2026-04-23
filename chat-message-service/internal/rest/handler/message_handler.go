package handler

import (
	"chat-message-service/internal/model"
	"chat-message-service/internal/service"
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type MessageHandler struct {
	service *service.MessageService
}

func NewMessageHandler(service *service.MessageService) *MessageHandler {
	return &MessageHandler{
		service: service,
	}
}

func (h *MessageHandler) RegisterRoutes(r chi.Router) {
	r.Route("/conversations/{id}/messages", func(r chi.Router) {
		r.Post("/", h.SendMessge)
		r.Get("/", h.GetConversationMessages)
		r.Get("/cursor", h.GetConversationMessagesCursor)
	})
}

func (h *MessageHandler) SendMessge(w http.ResponseWriter, r *http.Request) {
	var req model.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	//TODO: revisit
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing conversation id", http.StatusBadRequest)
		return
	}
	conversationID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "invalid conversation_id", http.StatusBadRequest)
		return
	}

	resp, err := h.service.Create(r.Context(), conversationID, req)
	if err != nil {
		http.Error(w, "failed to send message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)

}
func (h *MessageHandler) GetConversationMessages(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing conversation id", http.StatusBadRequest)
		return
	}
	conversationID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "invalid conversation_id", http.StatusBadRequest)
		return
	}

	// will remove once we have auth in place and can get user ID from token
	userId := r.URL.Query().Get("userid")
	if userId == "" {
		http.Error(w, "missing userid query parameter", http.StatusBadRequest)
		return
	}
	userUuid, err := uuid.Parse(userId)
	if err != nil {
		http.Error(w, "invalid userid query parameter", http.StatusBadRequest)
		return
	}

	messages, err := h.service.GetByConversation(r.Context(), userUuid, conversationID)
	if err != nil {
		http.Error(w, "failed to get messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)

}

func (h *MessageHandler) GetConversationMessagesCursor(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing conversation id", http.StatusBadRequest)
		return
	}
	conversationID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "invalid conversation_id", http.StatusBadRequest)
		return
	}

	// Default to MaxInt64 — returns the latest page when no cursor is provided
	beforeID := int64(math.MaxInt64)
	if raw := r.URL.Query().Get("mid"); raw != "" {
		mid, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			http.Error(w, "invalid 'before' id — must be an integer", http.StatusBadRequest)
			return
		}
		beforeID = mid
	}

	// mid := r.URL.Query().Get("mid")
	// if mid == "" {
	// 	http.Error(w, "missing message_id query parameter", http.StatusBadRequest)
	// 	return
	// }
	// messageID, err := strconv.ParseInt(mid, 10, 64)
	// if err != nil {
	// 	http.Error(w, "invalid message_id query parameter", http.StatusBadRequest)
	// 	return
	// }

	limitStr := r.URL.Query().Get("limit")
	limit := 20 // default limit
	if limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil || parsedLimit <= 0 {
			http.Error(w, "invalid limit query parameter", http.StatusBadRequest)
			return
		}
		limit = int(parsedLimit)
	}

	// will remove once we have auth in place and can get user ID from token
	userId := r.URL.Query().Get("userid")
	if userId == "" {
		http.Error(w, "missing userid query parameter", http.StatusBadRequest)
		return
	}
	userUuid, err := uuid.Parse(userId)
	if err != nil {
		http.Error(w, "invalid userid query parameter", http.StatusBadRequest)
		return
	}

	messages, err := h.service.GetByConversationCursor(r.Context(), userUuid, conversationID, beforeID, limit)
	if err != nil {
		http.Error(w, "failed to get messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)

}
