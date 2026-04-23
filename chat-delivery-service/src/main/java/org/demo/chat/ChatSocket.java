package org.demo.chat;

import java.util.Collections;
import java.util.HashSet;
import java.util.Set;
import java.util.stream.Collectors;

import org.demo.chat.client.CreateConversationRequest;
import org.demo.chat.client.CreateConversationResponse;
import org.demo.chat.client.MessageService;
import org.demo.chat.client.CreateConversationRequest.Participant;
import org.eclipse.microprofile.jwt.JsonWebToken;
import org.jboss.resteasy.reactive.ClientWebApplicationException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.quarkus.security.Authenticated;
import io.quarkus.websockets.next.CloseReason;
import io.quarkus.websockets.next.OnClose;
import io.quarkus.websockets.next.OnOpen;
import io.quarkus.websockets.next.OnTextMessage;
import io.quarkus.websockets.next.PathParam;
import io.quarkus.websockets.next.WebSocket;
import io.quarkus.websockets.next.WebSocketConnection;
import jakarta.inject.Inject;

/**
 * 
 * Represents ChatSocket.
 * 
 * The Jewel - connects users and other listeners (services)
 * through WS -> Broker
 *
 * @author A.Sabourov
 *
 */
@Authenticated
@WebSocket(path = "/chat/{username}")
public class ChatSocket {
	
	private static final Logger log = LoggerFactory.getLogger(ChatSocket.class);	

    @Inject
    ChatProducer producer;
    @Inject
    MessageService messageService;
    
    @Inject
    JsonWebToken jwt;
    
    /**
     * 
     * REVISIT:
     * The user parameter may go once auth tokens are implemented.
     * The user info would be taken from the token i.e. IDP sub claim
     * Which then can be used to lookup a logged in user through user interface (idp id -> local id)
     * Or as an alternative.. (less secure)
     * BFF actually can look it up via user service (idp id) which will return the logged in user id.
     * BFF would store it in session.
     * In that case UI can use it directly. 
     * 
     * 
     * @param connection
     * @param username - used for user uuid (revisit)
     * @return
     */
    @OnOpen
    public void onOpen(WebSocketConnection connection, @PathParam("username") String username) {
    	log.info("==> User joining: {}...", username);
    	
    	//TODO:
    	// get conversation IDs from Go message service that this
    	// user belongs to
    	// store:
    	
    	Set<Long> resp;
    	
    	try {
			resp = messageService.getConversationIDs(username);
		} catch (Exception e) {
			
			StringBuilder error = new StringBuilder();
			
			if (e instanceof ClientWebApplicationException) {
				error.append(((ClientWebApplicationException)e).getResponse().readEntity(String.class));
			} else {
				error.append(e.getMessage());
			}
			
	        log.error("Failed to get conversation IDs for user {}: {}", username, error);
	        connection.close(new CloseReason(4400, "Failed to initialize session due to: " + error))
		        .subscribe().with(
		                v -> log.info("Connection closed for user {} Error: {}", username, error),
		                err -> log.error("Failed to close connection: {}", err.getMessage())
	            );
	        log.info("Exiting...");
	        return;			
		}
    	
    	log.info("==> User joined: {}", username);
    	log.info("==> User's groups: {}", resp);
//    	Set<Long> groups = resp.stream().map(l -> String.valueOf(l)).collect(Collectors.toSet());
    	connection.userData().put(WsKeys.GROUPS, resp);
    	connection.userData().put(WsKeys.USERNAME, username);
    	
        connection.sendTextAndAwait(Message.builder()
                .type(MessageType.USER_JOINED)
                .from(username)
                .build());    	
    	
//    	return Message.builder()
//    			.type(MessageType.USER_JOINED)
//    			.from(username)
//    			.build();
    	
    }
    
//	@OnOpen
//	public Message onOpen(WebSocketConnection connection, @PathParam("username") String username) {    
//    	log.info("==> User joined: {}", username);
//    	
//    	connection.userData().put(WsKeys.USERNAME, username);
//    	
//    	return Message.builder()
//    			.type(MessageType.USER_JOINED)
//    			.from(username)
//    			.build();	
//	}
	
    @OnTextMessage
    public void onMessage(WebSocketConnection connection, Message message) {
    	
    	log.info("==> Message received: {}", message);
    	
    	if (message.to() == null) {
    		return;
    	}    	
    	
    	try {
    		// could be a new private chat. resolve (create if does not exists)
            producer.sendChat(resolvePrivateConversation(message, connection));
		} catch (Exception e) {
			// ignore for now
			log.warn("Error while sending a message: {} ", message, e);
		}
    	
    }

    @OnClose
    public void onClose(WebSocketConnection connection, @PathParam("username") String username) {
    	
    	log.info("==> User left: {}", username);
        
        Message departure = Message.builder()
    			.type(MessageType.USER_LEFT)
    			.from(username)
    			.build();
        connection.broadcast().sendTextAndAwait(departure);    	
    	
    }
    
    private Message resolvePrivateConversation(Message message, WebSocketConnection connection) throws Exception {
    	
    	if (message.conversationId() != null ) {
    		return message;
    	}
    	
    	if (message.to() == null || message.toName() == null) { 
    		throw new IllegalArgumentException("to and toName must be set for new private chat");
    	}
    	
    	if (message.from() == null || message.fromName() == null) { 
    		throw new IllegalArgumentException("from and fromName must be set for new private chat");
    	}
    	
    	Participant p1 = new Participant(message.from(), message.fromName());
    	Participant p2 = new Participant(message.to(), message.toName());
    	Set<Participant> participants = Set.of(p1, p2);
    	
    	CreateConversationRequest req = new CreateConversationRequest(null, "private", message.from(), participants);
    	
    	CreateConversationResponse resp = messageService.createConversation(req);
    	
    	// add private chat id for the user initiating the new private chat
    	connection.userData().get(WsKeys.GROUPS).add(resp.conversationId());
    	log.info("** Added user {} to private chat {}", message.from(), resp.conversationId());
    	
    	Message msg = Message.builder()
    			.id(message.id())
    			.type(message.type())
    			.conversationId(resp.conversationId())
    			.from(message.from())
    			.fromName(message.fromName())
    			.to(message.to())
    			.toName(message.toName())
    			.messageContent(message.message())
    			.build();
    			
    	return msg;
    			
    	
    }
}
