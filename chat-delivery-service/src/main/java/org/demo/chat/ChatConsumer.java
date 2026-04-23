package org.demo.chat;

import java.util.Set;

import org.eclipse.microprofile.reactive.messaging.Incoming;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.quarkus.websockets.next.OpenConnections;
import io.quarkus.websockets.next.WebSocketConnection;
import io.smallrye.mutiny.Uni;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;

/**
 * 
 * Represents ChatConsumer.
 * 
 * <p/>
 * 
 * Get a message from broker.
 * If the message is for a user that is connected (Web socket open) to
 * this service instance (replica in a cluster)
 * then deliver it to the user over WS connection.
 * Otherwise ignore it.
 *
 * @author A.Sabourov
 *
 */
@ApplicationScoped
public class ChatConsumer {
	
	private static final Logger log = LoggerFactory.getLogger(ChatConsumer.class);	
	
    @Inject
    OpenConnections connections;

    /**
     * first version user to user (from to to)
     * @param message
     * @return
     */
    //@Incoming("chat-in")
    public Uni<Void> consume(Message message) {
    	
    	log.info("==> Consumed: {}", message);
    	
        try {
            if (message.to() == null || message.to().isEmpty()) {
            	return Uni.createFrom().voidItem();
            }

            WebSocketConnection conn = get(message);
            if (conn != null) {
                conn.sendText(message).subscribe().with(
                        unused -> log.info("==> Sent to {} websocket: {}", message.to(), message),
                        failure -> log.error("==> Failed to send to {} websocket: {}", message.to(), message)
                    );
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
        
        return Uni.createFrom().voidItem();
    }
    
    /**
     * 
     * from user to group (public or private one on one )
     * 
     * @param message
     */
    @Incoming("chat-in")
    public void consumeConversaion(Message message) {
        // The "Target Room" for this specific message
        Long targetConversationId = message.conversationId();
        
        // First message for a new private chat
        if (message.to() != null) {
        	WebSocketConnection con = get(message);
        	if (con != null) {
        		con.userData().get(WsKeys.GROUPS).add(targetConversationId);
        		log.info("** Added user {} to private chat {}", message.to(), message.conversationId());
        	}
        }


        // 1. Get ALL active connections on THIS instance
        connections.stream()
            .filter(conn -> {
            	String user = conn.userData().get(WsKeys.USERNAME);
                // 2. Does THIS connected user belong to the TARGET room?
                Set<Long> userJoinedRooms = conn.userData().get(WsKeys.GROUPS);
                log.info("** User {} groups: {}", user, userJoinedRooms);
                return userJoinedRooms != null && userJoinedRooms.contains(targetConversationId);
            })
            .forEach(conn -> {
                // 3. If yes, deliver the message to this participant
                conn.sendText(message).subscribe().with(
                    unused -> log.info("==> Delivered to participant: {}", conn.userData().get(WsKeys.USERNAME)),
                    failure -> log.error("==> Delivery failed for user: {}", conn.id(), failure)
                );
            });
    }    
    
    private WebSocketConnection get(Message message) {
    	
    	for (WebSocketConnection c : connections) {
			if (message.to().equals(c.userData().get(WsKeys.USERNAME))) {
				return c;
			}
		}
    	log.debug("==> Target {} websocket not connected. Could be connected to another service instance.", message.to());
    	
    	return null;
    	
    }
}
