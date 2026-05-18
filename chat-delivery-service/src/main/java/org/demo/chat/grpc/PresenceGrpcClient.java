package org.demo.chat.grpc;

import org.demo.chat.grpc.ConnectionEvent;
import org.demo.chat.grpc.MutinyPresenceServiceGrpc;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.quarkus.grpc.GrpcClient;
import jakarta.enterprise.context.ApplicationScoped;

@ApplicationScoped
public class PresenceGrpcClient {

    private static final Logger log = LoggerFactory.getLogger(PresenceGrpcClient.class);

    @GrpcClient("presence-service")
    MutinyPresenceServiceGrpc.MutinyPresenceServiceStub stub;

    public void connected(String userUuid) {
        stub.connected(ConnectionEvent.newBuilder().setUserUuid(userUuid).build())
            .subscribe().with(
                ack -> log.debug("presence: connected {}", userUuid),
                err -> log.warn("presence connected failed for {}: {}", userUuid, err.getMessage())
            );
    }

    public void disconnected(String userUuid) {
        stub.disconnected(ConnectionEvent.newBuilder().setUserUuid(userUuid).build())
            .subscribe().with(
                ack -> log.debug("presence: disconnected {}", userUuid),
                err -> log.warn("presence disconnected failed for {}: {}", userUuid, err.getMessage())
            );
    }

    public void heartbeat(String userUuid) {
        stub.heartbeat(ConnectionEvent.newBuilder().setUserUuid(userUuid).build())
            .subscribe().with(
                ack -> {},
                err -> log.debug("presence heartbeat failed for {}: {}", userUuid, err.getMessage())
            );
    }
}
