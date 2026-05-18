package org.demo.chat.redis;

import java.util.Set;

import org.demo.chat.Message;
import org.demo.chat.MessageType;
import org.demo.chat.WsKeys;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.redis.datasource.ReactiveRedisDataSource;
import io.quarkus.redis.datasource.pubsub.ReactivePubSubCommands;
import io.quarkus.redis.datasource.pubsub.RedisPubSubMessage;
import io.quarkus.runtime.StartupEvent;
import io.quarkus.websockets.next.OpenConnections;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.enterprise.event.Observes;
import jakarta.inject.Inject;

@ApplicationScoped
public class TypingSubscriber {

    private static final Logger log = LoggerFactory.getLogger(TypingSubscriber.class);
    private static final String TYPING_PATTERN = "conversation:*:typing";
    private static final ObjectMapper MAPPER = new ObjectMapper();

    @Inject
    ReactiveRedisDataSource redis;

    @Inject
    OpenConnections connections;

    void onStart(@Observes StartupEvent event) {
        ReactivePubSubCommands<String> pubsub = redis.pubsub(String.class);
        pubsub.subscribeAsMessagesToPatterns(TYPING_PATTERN)
            .subscribe().with(
                msg -> handleTyping(msg.getChannel(), msg.getPayload()),
                err -> log.error("typing pub/sub subscribe failed: {}", err.getMessage())
            );
        log.info("subscribed to typing channels");
    }

    private void handleTyping(String channel, String payload) {
        // channel: conversation:{conversationId}:typing
        String[] parts = channel.split(":");
        if (parts.length < 3) return;

        long conversationId;
        try {
            conversationId = Long.parseLong(parts[1]);
        } catch (NumberFormatException e) {
            log.warn("unparseable conversation id in channel: {}", channel);
            return;
        }

        String userUuid;
        String username;
        try {
            JsonNode node = MAPPER.readTree(payload);
            userUuid = node.path("user_uuid").asText();
            username = node.path("username").asText(userUuid);
        } catch (Exception e) {
            log.warn("could not parse typing payload: {}", payload);
            return;
        }

        final long convId = conversationId;
        final String from = userUuid;
        final String fromName = username;

        Message typing = Message.builder()
            .type(MessageType.TYPING)
            .conversationId(convId)
            .from(from)
            .fromName(fromName)
            .build();

        connections.stream()
            .filter(conn -> {
                Set<Long> groups = conn.userData().get(WsKeys.GROUPS);
                return groups != null && groups.contains(convId);
            })
            .forEach(conn -> conn.sendText(typing)
                .subscribe().with(
                    v -> {},
                    err -> log.debug("typing fan-out failed: {}", err.getMessage())
                ));
    }
}
