package org.demo.chat.client;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.time.OffsetDateTime;

@JsonInclude(JsonInclude.Include.NON_NULL) // omit null fields in JSON
public record CreateConversationResponse(
        @JsonProperty("conversation_id") long conversationId,
        @JsonProperty("name") String name,                      // nullable
        @JsonProperty("type") String type,                      // "private" | "group"
        @JsonProperty("created_at") OffsetDateTime createdAt,
        @JsonProperty("last_message_at") OffsetDateTime lastMessageAt // nullable
) {}