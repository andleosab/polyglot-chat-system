package service

import (
	"chat-message-service/internal/kafka"
	"chat-message-service/internal/model"
	"chat-message-service/internal/repository"
	"context"
	"log"

	"github.com/google/uuid"
)

type ParticiopantService struct {
	participantRepo repository.ParticipantRepository
	producer        *kafka.Producer
}

func NewParticipantService(participantRepo repository.ParticipantRepository, producer *kafka.Producer) *ParticiopantService {
	return &ParticiopantService{
		participantRepo: participantRepo,
		producer:        producer,
	}
}

func (s *ParticiopantService) AddParticipant(ctx context.Context, conversationID int64, participant model.ParticipantRequest) (model.ParticipantResponse, error) {
	resp, err := s.participantRepo.AddParticipant(ctx, conversationID, participant)

	if err != nil {
		return model.ParticipantResponse{}, err
	}

	err = s.producer.PublishParticipantCreated(ctx, kafka.ParticipantEvent{
		ConversationID: conversationID,
		UserID:         participant.UserUUID.String(),
	})

	if err != nil {

		log.Println("==> Message not delivered", err)
		return model.ParticipantResponse{}, err
	}
	return resp, nil
}

func (s *ParticiopantService) GetConversationMembers(ctx context.Context, conversationID int64) ([]model.ParticipantResponse, error) {
	members, err := s.participantRepo.GetConversationMembers(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (s *ParticiopantService) GetGroups(ctx context.Context, userUUID string) ([]model.ConversationResponse, error) {
	resp, err := s.participantRepo.GetGroups(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// func (s *ParticiopantService) GetUserConversations(ctx context.Context, userUUID uuid.UUID) ([]model.UserConversationsResponse, error) {
// 	conversations, err := s.participantRepo.GetUserConversations(ctx, userUUID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return conversations, nil
// }

func (s *ParticiopantService) RemoveParticipant(ctx context.Context, conversationID int64, participantID uuid.UUID) error {
	err := s.participantRepo.RemoveParticipant(ctx, conversationID, participantID)
	if err != nil {
		return err
	}

	err = s.producer.PublishParticipantRemoved(ctx, kafka.ParticipantEvent{
		ConversationID: conversationID,
		UserID:         participantID.String(),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *ParticiopantService) UpdateAdminRole(ctx context.Context, conversationID int64, participantID uuid.UUID, isAdmin bool) error {
	return s.participantRepo.UpdateAdminRole(ctx, conversationID, participantID, isAdmin)
}
