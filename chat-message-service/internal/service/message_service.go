package service

import (
	"chat-message-service/internal/model"
	"chat-message-service/internal/repository"
	"context"

	"github.com/google/uuid"
)

type MessageService struct {
	repository repository.MessageRepository
}

func NewMessageService(repository repository.MessageRepository) *MessageService {
	return &MessageService{
		repository: repository,
	}
}

func (s MessageService) Create(ctx context.Context, conversationID int64, request model.MessageRequest) (model.MessageResponse, error) {
	resp, err := s.repository.Create(ctx, conversationID, request)
	if err != nil {
		return model.MessageResponse{}, err
	}
	return resp, nil
}

func (s MessageService) GetByConversation(ctx context.Context, userId uuid.UUID, conversationID int64) ([]model.MessageResponse, error) {
	resp, err := s.repository.GetByConversation(ctx, userId, conversationID, 100, 0)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s MessageService) GetByConversationCursor(ctx context.Context, userId uuid.UUID, conversationID int64, messageID int64, limit int) ([]model.MessageResponse, error) {
	resp, err := s.repository.GetByConversationCursor(ctx, userId, conversationID, messageID, limit)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s MessageService) Update(ctx context.Context, messageID int64, request model.MessageUpdateRequest) error {
	return s.repository.Update(ctx, messageID, request)
}

func (s MessageService) Delete(ctx context.Context, messageID int64) error {
	return s.repository.Delete(ctx, messageID)
}
