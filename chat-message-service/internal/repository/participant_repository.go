package repository

import (
	"chat-message-service/internal/model"
	"context"

	"github.com/google/uuid"
)

// ParticipantRepository defines the methods for participant-related operations
// using API-level model structs rather than SQLC structs.
type ParticipantRepository interface {
	// AddParticipant adds a new participant to a conversation
	AddParticipant(ctx context.Context, conversationID int64, participant model.ParticipantRequest) (model.ParticipantResponse, error)

	// GetConversationMembers retrieves all members of a conversation
	GetConversationMembers(ctx context.Context, conversationID int64) ([]model.ParticipantResponse, error)

	// GetUserConversations retrieves all conversations a user participates in
	// GetUserConversations(ctx context.Context, userUUID uuid.UUID) ([]model.UserConversationsResponse, error)

	// RemoveParticipant removes a participant from a conversation
	RemoveParticipant(ctx context.Context, conversationID int64, userUUID uuid.UUID) error

	// UpdateAdminRole updates a participant's admin role
	UpdateAdminRole(ctx context.Context, conversationID int64, userUUID uuid.UUID, isAdmin bool) error

	// GetGroups retrieves all groups a user is part of
	GetGroups(ctx context.Context, userUUID string) ([]model.ConversationResponse, error)
}
