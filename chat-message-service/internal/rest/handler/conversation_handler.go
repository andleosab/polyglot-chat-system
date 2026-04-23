package handler

import (
	"chat-message-service/internal/model"
	"chat-message-service/internal/rest/interceptor"
	"chat-message-service/internal/rest/response"
	"chat-message-service/internal/service"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Defines HTTP Handler for conversation-related endpoints.
type ConversationHandler struct {
	service *service.ConversationService
}

// NewConversationHandler creates a new instance of ConversationHandler.
func NewConversationHandler(service *service.ConversationService) *ConversationHandler {
	return &ConversationHandler{
		service: service,
	}
}

// Register relative routes for conversation endpoints.
func (h *ConversationHandler) RegisterRoutes(r chi.Router) {
	r.Route("/conversations", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)       // Chi only supports path routing (query params not supported)
		r.Get("/ids", h.listIds) // will be replaced by a gRPC endpoint in the future for intra-cluster communication})
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.Get)
			r.Put("/", h.Update)
			r.Delete("/", h.Delete)
		})
		r.Get("/private", h.GetPrivateConversation)
	})
}

func (h *ConversationHandler) Create(w http.ResponseWriter, r *http.Request) {

	var req model.ConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("==> Invalid body:", err)
		response.WriteJSONError(w, r, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		response.WriteJSONError(w, r, http.StatusBadRequest, "validation error: "+err.Error())
		return
	}

	log.Printf("==> Creating conversation with name: [%v], type: %s\n", req.Name, req.Type)

	if req.Type == "private" && (req.Name != nil) {
		response.WriteJSONError(w, r, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Type == "private" && (req.Participants == nil || len(req.Participants) != 2) {
		response.WriteJSONError(w, r, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Type == "group" && (req.Name == nil || *req.Name == "") {
		response.WriteJSONError(w, r, http.StatusBadRequest, "invalid request")
		return
	}

	conv, err := h.service.Create(r.Context(), req)
	if err != nil {
		log.Println("==> Failed to create conversation:", err)
		response.WriteJSONError(w, r, http.StatusInternalServerError, "failed to create conversation")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conv)

}

// Get returns a conversation by id
func (h *ConversationHandler) Get(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	conversationID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		response.WriteJSONError(w, r, http.StatusBadRequest, "invalid conversation id")
		return
	}

	conv, err := h.service.GetByID(r.Context(), conversationID)
	if err != nil {
		response.WriteJSONError(w, r, http.StatusInternalServerError, "failed to get conversation: "+err.Error())
		return
	}
	// Check if conversation was found
	if conv == (model.ConversationResponse{}) {
		response.WriteJSONError(w, r, http.StatusNotFound, "conversation not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conv)

}

// List returns all conversations for a user
func (h *ConversationHandler) List(w http.ResponseWriter, r *http.Request) {

	log.Println("==> Current user: ", interceptor.GetPrincipal(r.Context()))

	// REVISIT: get user ID from JWT claims once Auth is implemented
	// For now, we can get it from query params for testing purposes
	// This is not ideal but allows us to test the functionality without implementing JWT auth yet
	// uuid := common.CurrentUser(r).ID

	uuid := r.URL.Query().Get("id")
	if uuid == "" {
		response.WriteJSONError(w, r, http.StatusBadRequest, "missing user id")
		return
	}

	conversations, err := h.service.GetByUser(r.Context(), uuid)
	if err != nil {
		response.WriteJSONError(w, r, http.StatusInternalServerError, "failed to get conversations: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// ListIDs returns all conversation IDs for a user
// This is an optimization for the client to quickly fetch conversation IDs without loading all conversation details
// This will be replaced by a gRPC endpoint in the future for intra-cluster communication
func (h *ConversationHandler) listIds(w http.ResponseWriter, r *http.Request) {

	// REVISIT: get user ID from JWT claims once Auth is implemented
	// For now, we can get it from query params for testing purposes
	// This is not ideal but allows us to test the functionality without implementing JWT auth yet
	// uuid := common.CurrentUser(r).ID

	uuid := r.URL.Query().Get("id")
	if uuid == "" {
		response.WriteJSONError(w, r, http.StatusBadRequest, "missing user id")
		return
	}

	conversationIDs, err := h.service.GetConversationsIDs(r.Context(), uuid)
	if err != nil {
		response.WriteJSONError(w, r, http.StatusInternalServerError, "failed to get conversation ids: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversationIDs)

}

func (h *ConversationHandler) GetPrivateConversation(w http.ResponseWriter, r *http.Request) {

	user1 := r.URL.Query().Get("user1")
	user2 := r.URL.Query().Get("user2")

	if user1 == "" || user2 == "" {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	uuid1, err := uuid.Parse(user1)
	if err != nil {
		response.WriteJSONError(w, r, http.StatusBadRequest, "invalid input")
		return
	}
	uuid2, err := uuid.Parse(user2)
	if err != nil {
		response.WriteJSONError(w, r, http.StatusBadRequest, "invalid input")
		return
	}

	convID, err := h.service.GetPrivateConversation(r.Context(), uuid1, uuid2)

	if convID == 0 {
		response.WriteJSONError(w, r, http.StatusNotFound, "private conversation not found")
		return
	}

	if err != nil {
		response.WriteJSONError(w, r, http.StatusInternalServerError, "failed to get private conversation: "+err.Error())
		return
	}

	resp := model.ConversationLookupResponse{
		ConversationID: convID,
		Type:           "private",
	}

	response.WriteJSON(w, r, http.StatusOK, resp)
	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(resp)

}

func (h *ConversationHandler) Update(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	conversationID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		response.WriteJSONError(w, r, http.StatusBadRequest, "invalid conversation id")
		return
	}

	var req model.ConversationNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("==> Invalid body:", err)
		response.WriteJSONError(w, r, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}

	err = h.service.UpdateConversationName(r.Context(), conversationID, *req.Name)
	if err != nil {
		response.WriteJSONError(w, r, http.StatusInternalServerError,
			"failed to update conversation name: "+err.Error())
		return
	}
}

func (h *ConversationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// NOOP
}

/*
   REVISIT
   When sending message to a group the payload would have a conversation ID
   as the group is created already.

   When sending message to a user the payload would NOT have a conversation ID
   as the conversation does not exist yet.
   Instead, the payload should have the recipient user ID and the server will check if a conversation already exists between the sender and recipient.
   If it exists, the message will be sent to that conversation.
   If it does not exist, a new conversation will be created between the sender and recipient, and the message will be sent to that new conversation.

   This way we can support both group messages (with conversation ID) and direct messages (without conversation ID) in a flexible way.
*/
