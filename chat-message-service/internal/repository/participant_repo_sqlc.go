package repository

import (
	sqlcdb "chat-message-service/internal/db/sqlc"
	"chat-message-service/internal/model"
	"context"

	"github.com/google/uuid"
)

type participantRepoSQLC struct {
	q *sqlcdb.Queries
}

// Constructor
func NewParticipantRepoSQLC(q *sqlcdb.Queries) ParticipantRepository {
	return &participantRepoSQLC{q}
}

// Implement repository methods here...
func (r *participantRepoSQLC) AddParticipant(ctx context.Context, conversationID int64, p model.ParticipantRequest) (model.ParticipantResponse, error) {
	arg := sqlcdb.AddParticipantParams{
		ConversationID: conversationID,
		UserUuid:       p.UserUUID,
		Username:       p.Username,
		IsAdmin:        p.IsAdmin,
	}

	dbParticipant, err := r.q.AddParticipant(ctx, arg)
	if err != nil {
		return model.ParticipantResponse{}, err
	}

	return model.ParticipantResponse{
		ID:       dbParticipant.ID,
		UserUUID: dbParticipant.UserUuid,
		Username: dbParticipant.Username,
		IsAdmin:  dbParticipant.IsAdmin,
		JoinedAt: dbParticipant.JoinedAt.UTC(),
	}, nil
}

func (r *participantRepoSQLC) GetConversationMembers(ctx context.Context, conversationID int64) ([]model.ParticipantResponse, error) {
	rows, err := r.q.GetConversationMembers(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	var participants []model.ParticipantResponse
	for _, row := range rows {
		participants = append(participants, model.ParticipantResponse{
			ID:       row.ID,
			UserUUID: row.UserUuid,
			Username: row.Username,
			IsAdmin:  row.IsAdmin,
			JoinedAt: row.JoinedAt.UTC(),
		})
	}

	return participants, nil
}

// See ConversationRepoSQLC
// func (r *participantRepoSQLC) GetUserConversations(ctx context.Context, userUUID uuid.UUID) ([]model.UserConversationsResponse, error) {
// 	rows, err := r.q.GetUserConversations(ctx, userUUID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var convs []model.UserConversationsResponse
// 	for _, row := range rows {

// 		var lastMsg *time.Time
// 		if row.LastMessageAt.Valid {
// 			t := row.LastMessageAt.Time
// 			lastMsg = &t
// 		}

// 		convs = append(convs, model.UserConversationsResponse{
// 			ConversationID: row.ID,
// 			Name:           row.Name,
// 			Type:           string(row.Type),
// 			LastMessageAt:  lastMsg,
// 		})
// 	}

// 	return convs, nil
// }

func (r *participantRepoSQLC) RemoveParticipant(ctx context.Context, conversationID int64, userUUID uuid.UUID) error {
	arg := sqlcdb.RemoveParticipantParams{
		ConversationID: conversationID,
		UserUuid:       userUUID,
	}
	return r.q.RemoveParticipant(ctx, arg)
}

func (r *participantRepoSQLC) UpdateAdminRole(ctx context.Context, conversationID int64, userUUID uuid.UUID, isAdmin bool) error {
	arg := sqlcdb.UpdateAdminRoleParams{
		ConversationID: conversationID,
		UserUuid:       userUUID,
		IsAdmin:        isAdmin,
	}
	return r.q.UpdateAdminRole(ctx, arg)
}

func (r *participantRepoSQLC) GetGroups(ctx context.Context, userUUID string) ([]model.ConversationResponse, error) {
	uuid, err := uuid.Parse(userUUID) // validate UUID format early
	if err != nil {
		return nil, err
	}

	rows, err := r.q.GetUserGroups(ctx, uuid)
	if err != nil {
		return nil, err
	}

	groups := make([]model.ConversationResponse, 0, len(rows))
	for _, row := range rows {
		groups = append(groups, model.ConversationResponse{
			ConversationID: row.ID,
			Name:           row.GroupName,
			Type:           "group",
			CreatedAt:      row.CreatedAt,
		})
	}

	return groups, nil
}
