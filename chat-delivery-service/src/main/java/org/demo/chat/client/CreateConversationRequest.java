package org.demo.chat.client;

import java.util.Set;

import com.fasterxml.jackson.annotation.JsonProperty;

public record CreateConversationRequest(
		String name,
        String type, // "private" | "group"
        @JsonProperty("created_by")
        String createdBy,
        Set<Participant> participants		
) {
    public record Participant(
            @JsonProperty("user_uuid") 
            String userId,
            String username
    ) {}	
}
