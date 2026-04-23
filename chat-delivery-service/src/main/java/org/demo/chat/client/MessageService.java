package org.demo.chat.client;

import java.util.Set;

import org.demo.chat.config.ServiceConfig;
import org.eclipse.microprofile.rest.client.RestClientBuilder;
import org.eclipse.microprofile.rest.client.annotation.RegisterProvider;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import jakarta.ws.rs.Consumes;
import jakarta.ws.rs.GET;
import jakarta.ws.rs.POST;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.QueryParam;
import jakarta.ws.rs.core.MediaType;

@ApplicationScoped
public class MessageService {
	
	private static final Logger log = LoggerFactory.getLogger(MessageService.class);
	
	private final MessageClient messageClient;
	
	@Inject
	public MessageService(ServiceConfig serviceConfig) {
		this.messageClient = RestClientBuilder
				.newBuilder().baseUri(serviceConfig.messageServiceUrl())
				.register(getClass())
				.build(MessageClient.class);
	}
	
	/**
	 * 
	 * Revisit - change to gRpc
	 * or otherwise every call between services should also pass a JWT token with user id
	 * 
	 * @param uuid
	 * @return
	 */
	public Set<Long> getConversationIDs(String uuid) {
		return messageClient.getConversationIDs(uuid);
	}
	
	public CreateConversationResponse createConversation(CreateConversationRequest request) {
		return messageClient.createConversation(request);
	}
	
	
	@Path("/")
	@Consumes(MediaType.APPLICATION_JSON)
	@Produces(MediaType.APPLICATION_JSON)
	@RegisterProvider(ServiceAuthFilter.class)
	public static interface MessageClient {
		
		@GET
		@Path("/ids")
		Set<Long> getConversationIDs(@QueryParam("id") String uuid);
		
		@POST
		CreateConversationResponse createConversation(CreateConversationRequest request);
		
		
	}

}
