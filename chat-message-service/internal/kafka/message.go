package kafka

import (
	"log"

	"chat-message-service/internal/model"

	"github.com/google/uuid"
)

func MapKafkaMessage(kafkaMsg Message) (model.MessageRequest, error) {
	// parse sender UUID
	senderUUID, err := uuid.Parse(kafkaMsg.From)
	if err != nil {
		log.Printf("invalid sender UUID: %v, using Nil UUID", err)
		senderUUID = uuid.Nil
		return model.MessageRequest{}, err

	}

	// parse conversation ID
	// var convID int64
	// if kafkaMsg.ConversationID != "" {
	// 	convID, err = strconv.ParseInt(kafkaMsg.ConversationID, 10, 64)
	// 	if err != nil {
	// 		log.Printf("invalid conversation ID: %v, using 0", err)
	// 		convID = 0
	// 		return model.MessageRequest{}, err
	// 	}
	// }

	// map to internal model
	return model.MessageRequest{
		SenderUserUuid: senderUUID,
		ConversationID: kafkaMsg.ConversationID,
		Content:        kafkaMsg.Message,
		Timestamp:      kafkaMsg.Timestamp,
	}, nil
}

type Message struct {
	Type           string `json:"type"`
	ID             string `json:"id"`
	ConversationID int64  `json:"conversationid"`
	To             string `json:"to"`
	From           string `json:"from"`
	Message        string `json:"message"`
	Timestamp      int64  `json:"timestamp"`
}

type ParticipantEvent struct {
	// EventID        string `json:"eventId"` // EventID helps with idempotency downstream.
	ConversationID int64  `json:"conversationId"`
	UserID         string `json:"userId"`
	Timestamp      int64  `json:"timestamp"`
}
