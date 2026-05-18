package org.demo.chat.grpc;

import org.demo.chat.WsKeys;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.quarkus.scheduler.Scheduled;
import io.quarkus.websockets.next.OpenConnections;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;

@ApplicationScoped
public class HeartbeatScheduler {

    private static final Logger log = LoggerFactory.getLogger(HeartbeatScheduler.class);

    @Inject
    OpenConnections connections;

    @Inject
    PresenceGrpcClient presence;

    @Scheduled(every = "20s")
    void heartbeat() {
        connections.stream().forEach(conn -> {
            String userUuid = conn.userData().get(WsKeys.USERNAME);
            if (userUuid != null) {
                presence.heartbeat(userUuid);
            }
        });
        log.debug("heartbeat sent for {} connections", connections.stream().count());
    }
}
