package kafka

import (
	"chat-message-service/internal/config"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client *kgo.Client
	topics config.ProducerTopics
}

func NewProducer(ctx context.Context, config *config.Config) (*Producer, error) {

	log.Printf("Connecting Producer to Kafka brokers at %s", config.SeedBrokers)
	client, err := kgo.NewClient(
		kgo.SeedBrokers(config.SeedBrokers...),
		// reliability settings
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.RecordDeliveryTimeout(10*time.Second),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, err
	}

	return &Producer{
		client: client,
		topics: config.ProducerTopics,
	}, nil
}

func (p *Producer) PublishParticipantCreated(ctx context.Context, event ParticipantEvent) error {
	return p.publish(ctx, p.topics.ParticipantCreated, event)
}

func (p *Producer) PublishParticipantRemoved(ctx context.Context, event ParticipantEvent) error {
	return p.publish(ctx, p.topics.ParticipantRemoved, event)
}

func (p *Producer) publish(ctx context.Context, topic string, event ParticipantEvent) error {

	log.Println("==> Publishing to topic", p.topics.ParticipantCreated)

	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.client.ProduceSync(
		ctx,
		&kgo.Record{
			Topic: topic,
			Key: []byte(
				strconv.FormatInt(
					event.ConversationID,
					10,
				),
			),
			Value: value,
		},
	).FirstErr()
}

func (p *Producer) Close() {
	p.client.Close()
}
