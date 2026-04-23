// Provides application configuration by reading environment variables
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Config struct {
	DbUrl            string
	DbMaxConns       int32
	DbMaxIdleTimeSec int32
	HttpPort         string
	SeedBrokers      []string
	ConsumerGroup    string
	ConsumerTopic    string
	ProducerTopics
	JWTConfig
}

type ProducerTopics struct {
	ParticipantCreated string
	ParticipantRemoved string
}

type JWTConfig struct {
	Secret   []byte
	Issuer   []string
	Audience string
}

func LoadConfig() (*Config, error) {

	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/chat-message?sslmode=disable")

	httpPort := getEnv("HTTP_PORT", ":8080")
	if httpPort[0] != ':' {
		httpPort = ":" + httpPort
	}

	dbMaxConns, err := strconv.Atoi(getEnv("DB_MAX_CONNS", "0"))
	if err != nil {
		dbMaxConns = 0 // will use default in db package
	}
	dbMaxIdleTimeSec, err := strconv.Atoi(getEnv("DB_MAX_IDLE_TIME_SEC", "0"))
	if err != nil {
		dbMaxIdleTimeSec = 0 // will use default in db package
	}

	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	seedBrokers := strings.FieldsFunc(kafkaBrokers, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})

	consumerGroup := getEnv("KAFKA_CONSUMER_GROUP", "chat-message-service")
	consumerTopic := getEnv("KAFKA_CONSUMER_TOPIC", "chat.messages")

	secret, err := getRequiredEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	audience := getEnv("JWT_AUDIENCE", "chat-message-service")
	issuer, err := getRequiredEnv("JWT_ISSUER")
	if err != nil {
		return nil, err
	}

	// convert config.Audience comma-separated string to []string stripping whitespace
	issuers := strings.Split(issuer, ",")
	for i, a := range issuers {
		issuers[i] = strings.TrimSpace(a)
	}

	return &Config{
		DbUrl:            dbURL,
		HttpPort:         httpPort,
		DbMaxConns:       int32(dbMaxConns),
		DbMaxIdleTimeSec: int32(dbMaxIdleTimeSec),
		SeedBrokers:      seedBrokers,
		ConsumerGroup:    consumerGroup,
		ConsumerTopic:    consumerTopic,
		ProducerTopics: ProducerTopics{
			ParticipantCreated: getEnv("KAFKA_PRODUCER_TOPIC_PARTICIPANT_CREATED", "chat.participants.created"),
			ParticipantRemoved: getEnv("KAFKA_PRODUCER_TOPIC_PARTICIPANT_REMOVED", "chat.participants.removed"),
		},
		JWTConfig: JWTConfig{
			Secret:   []byte(secret),
			Issuer:   issuers,
			Audience: audience,
		},
	}, nil

}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getRequiredEnv(key string) (string, error) {
	if value := os.Getenv(key); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("required environment variable %q is not set", key)
}
