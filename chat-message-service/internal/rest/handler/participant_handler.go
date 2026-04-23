package handler

import (
	"chat-message-service/internal/model"
	"chat-message-service/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ParticipantHandler struct {
	service *service.ParticiopantService
}

func NewParticipantHandler(participantService *service.ParticiopantService) *ParticipantHandler {
	return &ParticipantHandler{
		service: participantService,
	}
}

// Register relative routes for conversation participant endpoints.
func (h *ParticipantHandler) RegisterRoutes(r chi.Router) {
	r.Route("/groups", func(r chi.Router) {
		r.Get("/", h.listGroups)
		r.Route("/{id}/participants", func(r chi.Router) {
			r.Post("/", h.AddParticipant)
			r.Get("/", h.GetConversationMembers)
			r.Delete("/{userId}", h.RemoveParticipant)
		})
	})
}

func (h *ParticipantHandler) listGroups(w http.ResponseWriter, r *http.Request) {

	uuid := r.URL.Query().Get("id")
	if uuid == "" {
		http.Error(w, "missing user id", http.StatusBadRequest)
		return
	}

	groups, err := h.service.GetGroups(r.Context(), uuid)
	if err != nil {
		http.Error(w, "failed to get groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

func (h *ParticipantHandler) AddParticipant(w http.ResponseWriter, r *http.Request) {
	var req model.ParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

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

	resp, err := h.service.AddParticipant(r.Context(), conversationID, req)
	if err != nil {
		http.Error(w, "failed to add participant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *ParticipantHandler) RemoveParticipant(w http.ResponseWriter, r *http.Request) {
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

	userId := chi.URLParam(r, "userId")
	if userId == "" {
		http.Error(w, "missing user id", http.StatusBadRequest)
		return
	}
	userID, err := uuid.Parse(userId)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	err = h.service.RemoveParticipant(r.Context(), conversationID, userID)
	if err != nil {
		http.Error(w, "failed to remove participant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ParticipantHandler) GetConversationMembers(w http.ResponseWriter, r *http.Request) {
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

	members, err := h.service.GetConversationMembers(r.Context(), conversationID)
	if err != nil {
		http.Error(w, "failed to get conversation members: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}
