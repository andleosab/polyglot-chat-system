package org.demo.chat;

import java.util.Objects;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

/**
 * 
 * Represents Message.
 *
 * @author A.Sabourov
 *
 */
@JsonIgnoreProperties(ignoreUnknown = true)
public record Message(
		MessageType type, 
		String id, 
		Long conversationId, 
		String to, 
		String toName,
		String from, 
		String fromName,
		String message, 
		long timestamp) {
	
    public static class Builder {
    	private String id;
        private MessageType type;
        private Long conversationId;
        private String to;
        private String toName;
        private String from;
        private String fromName;
        private String message;
        private long timestamp;

        public Builder type(MessageType type) {
            this.type = type;
            return this;
        }

        public Builder id(String id) {
            this.id = id;
            return this;
        }
        
        public Builder conversationId(Long conversationId) {
            this.conversationId = conversationId;
            return this;
        }
        
        public Builder to(String to) {
            this.to = to;
            return this;
        }

        public Builder toName(String toName) {
            this.toName = toName;
            return this;
        }

        public Builder from(String from) {
            this.from = from;
            return this;
        }

        public Builder fromName(String fromName) {
            this.fromName = fromName;
            return this;
        }

        public Builder messageContent(String message) {
            this.message = message;
            return this;
        }

        public Builder timestamp(long timestamp) {
            this.timestamp = timestamp;
            return this;
        }
        
        public Message build() {
//            Objects.requireNonNull(type, "Type cannot be null");
//            Objects.requireNonNull(to, "To cannot be null");
//            Objects.requireNonNull(from, "From cannot be null");
//            Objects.requireNonNull(message, "Message content cannot be null");

            return new Message(type, id, conversationId, to, toName, 
            		from, fromName, message, timestamp);
        }
    }

    public static Builder builder() {
        return new Builder();
    }

}
