package org.demo.chat;

import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import org.eclipse.microprofile.reactive.messaging.Channel;
import org.eclipse.microprofile.reactive.messaging.Emitter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * 
 * Represents ChatProducer.
 * 
 * Pretty clear - emit a message to a broker.
 *
 * @author A.Sabourov
 *
 */
@ApplicationScoped
public class ChatProducer {
	
	private static final Logger log = LoggerFactory.getLogger(ChatProducer.class);	
	
    @Inject
    @Channel("chat-out")
    Emitter<Message> chatEmitter;

    public void sendChat(Message msg) {
        chatEmitter.send(msg);
        log.info("==> Produced: {}", msg);
    }
}
