package repository

import (
	sqlcdb "chat-message-service/internal/db/sqlc"
	"chat-message-service/internal/model"
	"database/sql"
	"errors"
	"log"

	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type conversationRepoSQLC struct {
	q    *sqlcdb.Queries
	pool *pgxpool.Pool
}

// this retunrs the interface type and so it actually a pointer to the struct implementing the interface
func NewConversationRepoSQLC(q *sqlcdb.Queries, pool *pgxpool.Pool) ConversationRepository {
	return &conversationRepoSQLC{q, pool}
}

// using pointer for name to allow null values (Go specific handling of optional fields and nullability in DB)
func (r *conversationRepoSQLC) Create(ctx context.Context, conversation model.ConversationRequest) (model.ConversationResponse, error) {

	typ := sqlcdb.ConversationType(conversation.Type)

	var resp model.ConversationResponse

	switch typ {
	case sqlcdb.ConversationTypeGroup:
		c, err := r.q.CreateConversation(ctx, sqlcdb.CreateConversationParams{
			Type:      typ,
			Name:      conversation.Name,
			CreatedBy: conversation.CreatedBy,
		})
		if err != nil {
			return model.ConversationResponse{}, err
		}
		resp = toModel(c)
	case sqlcdb.ConversationTypePrivate:
		tx, err := r.pool.Begin(ctx)

		log.Println("In transaction...")

		if err != nil {
			return model.ConversationResponse{}, err
		}
		defer tx.Rollback(ctx)
		qtx := r.q.WithTx(tx)

		//0. check if private conversation already exists between the 2 participants
		existingID, err := r.lookupPrivateConversationID(ctx, conversation.Participants)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return model.ConversationResponse{}, err
		}
		if existingID != 0 {
			existingConv, err := qtx.GetConversationByID(ctx, existingID)
			if err != nil {
				return model.ConversationResponse{}, err
			}
			return toModel(existingConv), nil
		}

		// 1. create conversation
		c, err := qtx.CreateConversation(ctx, sqlcdb.CreateConversationParams{
			Type:      typ,
			Name:      nil,
			CreatedBy: conversation.CreatedBy,
		})
		if err != nil {
			return model.ConversationResponse{}, err
		}

		// 2. add participants
		for _, p := range conversation.Participants {
			_, err = qtx.AddParticipant(ctx, sqlcdb.AddParticipantParams{
				ConversationID: c.ID,
				UserUuid:       p.UserUUID,
				Username:       p.Username,
			})
			if err != nil {
				return model.ConversationResponse{}, err
			}
		}

		err = tx.Commit(ctx)
		if err != nil {
			return model.ConversationResponse{}, err
		}

		resp = toModel(c)
	default:
		return model.ConversationResponse{}, errors.New("invalid conversation type")
	}

	return resp, nil
}

// GetByID returns an error if no rows found triggered by query definition that uses :one
// So an empty struct is returned if not found instead of an error
func (r *conversationRepoSQLC) GetByID(ctx context.Context, conversationID int64) (model.ConversationResponse, error) {
	c, err := r.q.GetConversationByID(ctx, conversationID)
	if err != nil {
		// Handle sql.ErrNoRows to return a nil model and no error if not found
		if errors.Is(err, sql.ErrNoRows) {
			return model.ConversationResponse{}, nil
		}
		return model.ConversationResponse{}, err
	}
	return toModel(c), nil
}

func (r *conversationRepoSQLC) UpdateName(ctx context.Context, conversationID int64, newName string) error {
	params := sqlcdb.UpdateConversationNameParams{
		Name: &newName,
		ID:   conversationID,
	}
	return r.q.UpdateConversationName(ctx, params)
}

func (r *conversationRepoSQLC) DeleteByUUID(ctx context.Context, conversationID int64) error {
	return r.q.DeleteConversation(ctx, conversationID)
}

func (r *conversationRepoSQLC) UpdateConversationLastMessageAt(ctx context.Context, conversationID int64) error {
	return r.q.UpdateConversationLastMessageAt(ctx, conversationID)
}

func (r *conversationRepoSQLC) GetConversationsIDs(ctx context.Context, userUUID string) ([]int64, error) {
	uuid, err := uuid.Parse(userUUID) // validate UUID format early
	if err != nil {
		return nil, err
	}

	ids, err := r.q.GetUserConversationIDs(ctx, uuid)
	if err != nil {
		return nil, err
	}

	if ids == nil {
		return []int64{}, nil
	}
	return ids, nil
}

// GetConversations returns all conversations (chats) for a user
func (r *conversationRepoSQLC) GetConversations(ctx context.Context, userUUID string) ([]model.ConversationResponse, error) {
	uuid, err := uuid.Parse(userUUID) // validate UUID format early
	if err != nil {
		return nil, err
	}

	rows, err := r.q.GetUserConversations(ctx, uuid)
	if err != nil {
		log.Println("==> Error occurred while fetching user conversations:", err)
		return nil, err
	}

	// var conversations []model.ConversationResponse
	// Pre-allocate capacity but keep length at 0
	//TODO: change to use offset and limit later
	conversations := make([]model.ConversationResponse, 0, len(rows))
	for _, row := range rows {
		conversations = append(conversations, rowToModel(row))
	}

	return conversations, nil

}

func (r *conversationRepoSQLC) GetPrivateConversationID(ctx context.Context, userUUID1, userUUID2 uuid.UUID) (int64, error) {
	return r.q.GetPrivateConversationID(ctx, sqlcdb.GetPrivateConversationIDParams{
		UserUuid1: userUUID1,
		UserUuid2: userUUID2,
	})
}

func (r *conversationRepoSQLC) lookupPrivateConversationID(ctx context.Context, participants []model.Participant) (int64, error) {
	return r.q.GetPrivateConversationID(ctx, sqlcdb.GetPrivateConversationIDParams{
		UserUuid1: participants[0].UserUUID,
		UserUuid2: participants[1].UserUUID,
	})
}

func (r *conversationRepoSQLC) UpdateConversationName(ctx context.Context, conversationID int64, newName string) error {
	return r.q.UpdateConversationName(ctx, sqlcdb.UpdateConversationNameParams{
		Name: &newName,
		ID:   conversationID,
	})
}

func rowToModel(c sqlcdb.GetUserConversationsRow) model.ConversationResponse {

	// var lastMessageAt *time.Time // nil by default
	// if c.SentAt.Valid {
	// 	lastMsg := c.SentAt.Time // intermediate variable
	// 	lastMessageAt = &lastMsg
	// }

	model := model.ConversationResponse{
		ConversationID: c.ID,
		Name:           c.DisplayName,
		Type:           string(c.Type),
		CreatedAt:      c.CreatedAt.UTC(),
		LastMessageAt:  toUtcTime(c.SentAt),
	}

	return model

}

func toModel(c sqlcdb.Conversation) model.ConversationResponse {

	// var lastMessageAt *time.Time // nil by default
	// if c.LastMessageAt.Valid {
	// 	lastMsg := c.LastMessageAt.Time // intermediate variable
	// 	lastMessageAt = &lastMsg
	// }

	model := model.ConversationResponse{
		ConversationID: c.ID,
		Name:           c.Name,
		Type:           string(c.Type),
		CreatedAt:      c.CreatedAt.UTC(),
		CreatedBy:      c.CreatedBy,
		LastMessageAt:  toUtcTimePtr(c.LastMessageAt),
	}

	return model
}
