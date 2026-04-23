package model

import (
	"time"

	"github.com/google/uuid"
)

// ParticipantRequest is used when adding a participant to a conversation
type ParticipantRequest struct {
	UserUUID uuid.UUID `json:"user_uuid" validate:"required"`
	Username string    `json:"username" validate:"required"`
	IsAdmin  bool      `json:"is_admin"`
}

// ParticipantResponse is returned in API responses for conversation members
type ParticipantResponse struct {
	ID       int64     `json:"id"`
	UserUUID uuid.UUID `json:"user_uuid"`
	Username string    `json:"username"`
	IsAdmin  bool      `json:"is_admin"`
	JoinedAt time.Time `json:"joined_at"`
}

// UserConversationsResponse represents a conversation for a particular user
type UserConversationsResponse struct {
	ConversationID int64      `json:"conversation_uuid"`
	Name           string     `json:"name"`
	Type           string     `json:"type"`
	LastMessageAt  *time.Time `json:"last_message_at,omitempty"`
}
