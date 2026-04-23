package repository

import (
	"chat-message-service/internal/model"

	"context"

	"github.com/google/uuid"
)

// ConversationRepository defines DB operations for conversations.
type ConversationRepository interface {
	Create(ctx context.Context, conversation model.ConversationRequest) (model.ConversationResponse, error)
	GetByID(ctx context.Context, conversationID int64) (model.ConversationResponse, error)
	UpdateName(ctx context.Context, conversationID int64, newName string) error
	DeleteByUUID(ctx context.Context, conversationID int64) error
	UpdateConversationLastMessageAt(ctx context.Context, conversationID int64) error
	// GetConversations(ctx context.Context, userUUID string, limit int, offset int) ([]model.ConversationResponse, error)
	GetConversations(ctx context.Context, userUUID string) ([]model.ConversationResponse, error)
	GetConversationsIDs(ctx context.Context, userUUID string) ([]int64, error)
	GetPrivateConversationID(ctx context.Context, userUUID1, userUUID2 uuid.UUID) (int64, error)
	UpdateConversationName(ctx context.Context, conversationID int64, newName string) error
}
