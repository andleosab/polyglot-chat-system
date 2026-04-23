package org.demo.chat.client;

import java.time.Duration;

import javax.crypto.SecretKey;

import org.demo.chat.config.ServiceConfig;
import org.eclipse.microprofile.jwt.JsonWebToken;

import io.quarkus.security.identity.SecurityIdentity;
import io.smallrye.jwt.algorithm.SignatureAlgorithm;
import io.smallrye.jwt.build.Jwt;
import io.smallrye.jwt.util.KeyUtils;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;

@ApplicationScoped
public class ServiceTokenFactory {

    @Inject
    ServiceConfig serviceConfig;
    
    @Inject
    SecurityIdentity identity;    

    public String generate() {
        SecretKey key = KeyUtils.createSecretKeyFromSecret(serviceConfig.jwtSecret());
        
        JsonWebToken jwt = (JsonWebToken) identity.getPrincipal();
        String userId = jwt.getSubject();

        return Jwt.claims()
            .issuer(serviceConfig.jwtIssuer())
            .subject(userId)
            .audience("chat-message-service")
            .expiresIn(Duration.ofSeconds(30))
            .jws()
            .algorithm(SignatureAlgorithm.HS256)
            .sign(key);
    }
}