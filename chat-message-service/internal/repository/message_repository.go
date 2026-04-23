package repository

import (
	"chat-message-service/internal/model"
	"context"

	"github.com/google/uuid"
)

type MessageRepository interface {
	Create(ctx context.Context, conversationID int64, message model.MessageRequest) (model.MessageResponse, error)
	GetByConversation(ctx context.Context, userId uuid.UUID, conversationID int64, limit int, offset int) ([]model.MessageResponse, error)
	GetByConversationCursor(ctx context.Context, userId uuid.UUID, conversationID int64, messageID int64, limit int) ([]model.MessageResponse, error)
	Update(ctx context.Context, messageID int64, message model.MessageUpdateRequest) error
	Delete(ctx context.Context, messageID int64) error
}
