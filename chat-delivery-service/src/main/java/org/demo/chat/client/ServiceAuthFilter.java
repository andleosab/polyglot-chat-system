package org.demo.chat.client;

import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import jakarta.ws.rs.client.ClientRequestContext;
import jakarta.ws.rs.client.ClientRequestFilter;
import jakarta.ws.rs.ext.Provider;

@ApplicationScoped
@Provider
public class ServiceAuthFilter implements ClientRequestFilter {

    @Inject
    ServiceTokenFactory tokenFactory;

    @Override
    public void filter(ClientRequestContext ctx) {
    	ctx.getHeaders().add("Authorization", "Bearer " + tokenFactory.generate());
    }
}