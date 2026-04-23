package org.demo.chat;

public record ParticipantEvent(
		Long conversationId,
		String userId,
		Long timestamp
) {}
