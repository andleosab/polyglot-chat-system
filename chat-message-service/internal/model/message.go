package model

import (
	"time"

	"github.com/google/uuid"
)

type MessageRequest struct {
	ConversationID int64     `json:"conversation_id" validate:"required"`
	SenderUserUuid uuid.UUID `json:"sender_uuid" validate:"required"`
	Content        string    `json:"content" validate:"required"`
	Timestamp      int64     `json:"timestamp"`
}

type MessageResponse struct {
	ID             int64      `json:"id"`
	ConversationID int64      `json:"conversation_id"`
	SenderUserUuid *uuid.UUID `json:"sender_uuid,omitempty"`
	SenderUsername string     `json:"sender_name,omitempty"`
	Content        string     `json:"content"`
	SentAt         time.Time  `json:"sent_at"`
	EditedAt       *time.Time `json:"edited_at,omitempty"`
	Timestamp      *int64     `json:"timestamp"`
}

type MessageUpdateRequest struct {
	ID      int64  `json:"id" validate:"required"`
	Content string `json:"content" validate:"required"`
}
