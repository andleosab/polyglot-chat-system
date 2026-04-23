package service

import (
	"chat-message-service/internal/model"
	"chat-message-service/internal/repository"
	"context"
	"log"

	"github.com/google/uuid"
)

// ConversationService provides operations related to conversations.
type ConversationService struct {
	repository repository.ConversationRepository
}

// NewConversationService creates a new instance of ConversationService.
// the argument is an interface type so it's a pointer to the struct implementing the interface
func NewConversationService(repository repository.ConversationRepository) *ConversationService {
	return &ConversationService{
		repository: repository,
	}
}

// Create creates a new conversation API model.
func (s *ConversationService) Create(ctx context.Context, conversation model.ConversationRequest) (model.ConversationResponse, error) {

	resp, err := s.repository.Create(ctx, conversation)
	if err != nil {
		return model.ConversationResponse{}, err
	}

	return resp, nil
}

// GetByID returns a conversation by id
func (s *ConversationService) GetByID(ctx context.Context, conversationID int64) (model.ConversationResponse, error) {

	resp, err := s.repository.GetByID(ctx, conversationID)
	if err != nil {
		return model.ConversationResponse{}, err
	}

	return resp, nil
}

// GetByUser returns a list of conversations for a user
func (s *ConversationService) GetByUser(ctx context.Context, userUUID string) ([]model.ConversationResponse, error) {
	resp, err := s.repository.GetConversations(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetConversationsIDs returns a list of conversation IDs for a user
func (s *ConversationService) GetConversationsIDs(ctx context.Context, userUUID string) ([]int64, error) {
	resp, err := s.repository.GetConversationsIDs(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ConversationService) GetPrivateConversation(ctx context.Context, userUUID1, userUUID2 uuid.UUID) (int64, error) {
	resp, err := s.repository.GetPrivateConversationID(ctx, userUUID1, userUUID2)

	if err != nil {
		log.Println("Error (if any) while fetching private conversation ID:", err)
		return 0, err
	}
	log.Println("Private conversation ID:", resp)

	return resp, nil
}

func (s *ConversationService) UpdateConversationName(ctx context.Context, conversationID int64, newName string) error {
	return s.repository.UpdateConversationName(ctx, conversationID, newName)
}

// moved to repository/conversation_repo_sqlc.go
// func toModel(c sqlcdb.Conversation) *model.ConversationResponse {

// 	model := &model.ConversationResponse{
// 		ConvewrsationUUID: c.ConversationUuid.String(),
// 		Name:              &c.Name.String,
// 		Type:              string(c.Type),
// 		CreatedAt:         c.CreatedAt.Time,
// 		LastMessageAt:     &c.LastMessageAt.Time,
// 	}

// 	// the using intermediate variable pattern avoids hodling a reference to the source struct (db struct in this case)
// 	// thus allowing garbage collection to free up memory if needed sooner
// 	if c.Name.Valid {
// 		name := c.Name.String // intermediate variable
// 		model.Name = &name
// 	}

// 	if c.LastMessageAt.Valid {
// 		lastMessageAt := c.LastMessageAt.Time // intermediate variable
// 		model.LastMessageAt = &lastMessageAt
// 	}

// 	return model
// }
