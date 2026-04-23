package repository

import (
	sqlcdb "chat-message-service/internal/db/sqlc"
	"chat-message-service/internal/model"
	"context"
	"log"

	"github.com/google/uuid"
)

type MessageRepoSQLC struct {
	*sqlcdb.Queries
}

func NewMessageRepoSQLC(queries *sqlcdb.Queries) MessageRepository {
	return &MessageRepoSQLC{queries}
}

func (r *MessageRepoSQLC) Create(ctx context.Context, conversationID int64,
	message model.MessageRequest) (model.MessageResponse, error) {
	msg, err := r.Queries.CreateMessage(ctx, sqlcdb.CreateMessageParams{
		ConversationID: conversationID,
		SenderUserUuid: message.SenderUserUuid,
		Content:        message.Content,
		ClientTs:       &message.Timestamp,
	})
	if err != nil {
		return model.MessageResponse{}, err
	}
	return toModelCreateMessage(msg), nil
}
func (r *MessageRepoSQLC) GetByConversation(ctx context.Context, userId uuid.UUID, conversationID int64, limit int, offset int) ([]model.MessageResponse, error) {
	msgs, err := r.Queries.GetMessagesByConversation(ctx, sqlcdb.GetMessagesByConversationParams{
		ConversationID: conversationID,
		Limit:          int32(limit),
		Offset:         int32(offset),
		UserUuid:       userId,
	})
	if err != nil {
		return nil, err
	}

	log.Println("--> Looking up messages by conversation: ", conversationID)

	var result []model.MessageResponse = []model.MessageResponse{}
	for _, m := range msgs {
		result = append(result, toModelConversationMessage(m))
	}
	return result, nil
}

func (r *MessageRepoSQLC) GetByConversationCursor(ctx context.Context, userId uuid.UUID, conversationID int64, messageID int64, limit int) ([]model.MessageResponse, error) {
	msgs, err := r.Queries.GetMessagesByConversationCursor(ctx, sqlcdb.GetMessagesByConversationCursorParams{
		UserUuid:       userId,
		ConversationID: conversationID,
		MessageID:      messageID,
		Limit:          int32(limit),
	})
	if err != nil {
		return nil, err
	}

	log.Printf("--> Looking up messages by conversation: %d with cursor message ID: %d\n", conversationID, messageID)

	var result []model.MessageResponse = []model.MessageResponse{}
	for _, m := range msgs {
		result = append(result, toModelConversationMessageCursor(m))
	}
	return result, nil
}

func (r *MessageRepoSQLC) Update(ctx context.Context, messageID int64, message model.MessageUpdateRequest) error {
	return r.Queries.UpdateMessageContent(ctx, sqlcdb.UpdateMessageContentParams{
		Content: message.Content,
		ID:      messageID,
	})
}
func (r *MessageRepoSQLC) Delete(ctx context.Context, messageID int64) error {
	return r.Queries.DeleteMessage(ctx, messageID)
}

func toModelCreateMessage(m sqlcdb.Message) model.MessageResponse {
	return model.MessageResponse{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderUserUuid: &m.SenderUserUuid,
		Content:        m.Content,
		SentAt:         m.SentAt.UTC(),
		EditedAt:       toUtcTimePtr(m.EditedAt),
		Timestamp:      m.ClientTs,
	}
}

func toModelConversationMessage(m sqlcdb.GetMessagesByConversationRow) model.MessageResponse {
	return model.MessageResponse{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderUserUuid: &m.SenderUserUuid,
		SenderUsername: m.SenderUsername,
		Content:        m.Content,
		SentAt:         m.SentAt.UTC(),
		EditedAt:       toUtcTimePtr(m.EditedAt),
		Timestamp:      m.ClientTs,
	}
}

func toModelConversationMessageCursor(m sqlcdb.GetMessagesByConversationCursorRow) model.MessageResponse {
	return model.MessageResponse{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderUserUuid: &m.SenderUserUuid,
		SenderUsername: m.SenderUsername,
		Content:        m.Content,
		SentAt:         m.SentAt.UTC(),
		EditedAt:       toUtcTimePtr(m.EditedAt),
		Timestamp:      m.ClientTs,
	}
}
