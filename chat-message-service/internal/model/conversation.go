package model

import (
	"time"

	"github.com/google/uuid"
)

type Participant struct {
	UserUUID uuid.UUID `json:"user_uuid" validate:"required"`
	Username string    `json:"username" validate:"required"`
}

// TODO: Created_By may be removed once Auth2/OIDC Bearer token is implemented
// move to a separate file
type ConversationRequest struct {
	Name         *string       `json:"name,omitempty"`
	Type         string        `json:"type" validate:"oneof=private group"`
	CreatedBy    uuid.UUID     `json:"created_by" validate:"required"`
	Participants []Participant `json:"participants,omitempty" validate:"omitempty,min=2,max=2,dive,required"`
}

type ConversationNameRequest struct {
	Name *string `json:"name" validate:"required, min=3, max=50"`
}

type ConversationResponse struct {
	ConversationID int64      `json:"conversation_id"`
	Name           *string    `json:"name,omitempty"`
	Type           string     `json:"type" validate:"oneof=private group"`
	CreatedAt      time.Time  `json:"created_at"`
	CreatedBy      uuid.UUID  `json:"created_by"`
	LastMessageAt  *time.Time `json:"last_message_at,omitempty"`
}

type ConversationLookupResponse struct {
	ConversationID int64  `json:"conversation_id"`
	Type           string `json:"type" validate:"oneof=private group"`
}
