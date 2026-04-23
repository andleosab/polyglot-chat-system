package kafka

import (
	"chat-message-service/internal/config"
	"chat-message-service/internal/repository"
	"context"
	"log"

	"encoding/json"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Consumer struct {
	client *kgo.Client
	repo   repository.MessageRepository
}

func NewConsumer(ctx context.Context, config *config.Config, repo repository.MessageRepository) (*Consumer, error) {

	log.Printf("Connecting to Kafka brokers at %s", config.SeedBrokers)
	client, err := kgo.NewClient(
		kgo.SeedBrokers(config.SeedBrokers...),
		kgo.ConsumerGroup(config.ConsumerGroup),
		kgo.ConsumeTopics(config.ConsumerTopic),
		// kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()), // read from beginning if needed
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, err
	}

	return &Consumer{
		client: client,
		repo:   repo,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	consume(ctx, c.client, c.repo)
}

func (c *Consumer) Stop() {
	c.client.Close()
}

func consume(ctx context.Context, client *kgo.Client, repo repository.MessageRepository) {

	for {
		// fetches := c.PollFetches(context.Background()) // exits when client is closed
		fetches := client.PollFetches(ctx) // exists faster when using app context
		if fetches.IsClientClosed() {
			return
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, err := range errs {
				if err.Err == context.Canceled {
					continue // don't log it
				}
				log.Printf(
					"topic=%s partition=%d err=%v",
					err.Topic,
					err.Partition,
					err.Err,
				)
			}
			continue
		}

		const maxBatch = 100
		var processed []*kgo.Record

		log.Println("Start processing messages...")
		for _, record := range fetches.Records() {

			log.Printf("==> Received message: %s", record.Value)

			var kafkaMsg Message
			if err := json.Unmarshal(record.Value, &kafkaMsg); err != nil {
				log.Printf("failed to unmarshal message. Skipping it: %v", err)

				client.CommitRecords(ctx, record) // skip poison message
				continue
			}

			log.Printf(
				"conversationId=%d senderId=%s text=%s",
				kafkaMsg.ConversationID,
				kafkaMsg.From,
				kafkaMsg.Message,
			)

			chatMsg, err := MapKafkaMessage(kafkaMsg)
			if err != nil {
				log.Printf("failed to map Kafka message: %v", err)
				client.CommitRecords(ctx, record)
				continue
			}

			// log.Printf("==> chatMsg: conversationId=%d senderUserUuid=%s content=%s",
			// 	chatMsg.ConversationID,
			// 	chatMsg.SenderUserUuid,
			// 	chatMsg.Content,
			// )

			// save record to DB
			if _, err := repo.Create(ctx, chatMsg.ConversationID, chatMsg); err != nil {
				log.Printf("failed to save message to DB. Skipping it: %v", err)
				client.CommitRecords(ctx, record)
				continue
			}

			log.Printf("==> Processed message: %s", record.Value)

			processed = append(processed, record)
			if len(processed) >= maxBatch {
				client.CommitRecords(ctx, processed...) // commit batch
				processed = processed[:0]               // reset slice
			}

		}

		// commit any remaining records after loop
		if len(processed) > 0 {
			client.CommitRecords(ctx, processed...)
		}

	}

}
