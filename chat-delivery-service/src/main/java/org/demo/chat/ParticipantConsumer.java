package org.demo.chat;

import java.util.Set;

import org.eclipse.microprofile.reactive.messaging.Incoming;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.quarkus.websockets.next.OpenConnections;
import io.quarkus.websockets.next.WebSocketConnection;
import jakarta.inject.Inject;

public class ParticipantConsumer {
	
	private static final Logger log = LoggerFactory.getLogger(ParticipantConsumer.class);	
	
    @Inject
    OpenConnections connections;
	
    @Incoming("participant-created")
    public void participantCreated(ParticipantEvent event) {
    	
    	WebSocketConnection con = get(event.userId());
    	
    	if (con != null) {
    		Set<Long> groups = con.userData().get(WsKeys.GROUPS);
    		if (groups != null) {
        		con.userData().put(WsKeys.GROUPS, groups);
        		groups.add(event.conversationId());
        		log.info("** Added online user {} to group chat {}", event.userId(), event.conversationId());
    		}
    	}    	
    	
    }

    @Incoming("participant-removed")
    public void participantRemoved(ParticipantEvent event) {
    	
    	WebSocketConnection con = get(event.userId());
    	
    	if (con != null) {
    		Set<Long> groups = con.userData().get(WsKeys.GROUPS);
    		if (groups != null) {
        		con.userData().get(WsKeys.GROUPS).remove(event.conversationId());
        		log.info("** Removed online user {} from group chat {}", event.userId(), event.conversationId());
    		}
    	}    	
    }
    
    private WebSocketConnection get(String userId) {
    	
    	for (WebSocketConnection c : connections) {
			if (userId.equals(c.userData().get(WsKeys.USERNAME))) {
				return c;
			}
		}
    	return null;
    	
    }    
}
