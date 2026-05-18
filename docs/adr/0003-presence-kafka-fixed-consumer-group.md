# ADR-0003: Presence service has no Kafka consumer

## Status
Superseded by ADR-0004

## Context
The original presence design proposed a `chat.connections` Kafka topic and a fixed consumer group (`chat-presence-service`) for the presence service to consume connection events.

## Decision
This ADR is superseded by ADR-0004. All delivery→presence signals (`Connected`, `Disconnected`, `Heartbeat`) were moved to gRPC direct calls, eliminating the `chat.connections` Kafka topic entirely. The presence service has no Kafka consumer.
