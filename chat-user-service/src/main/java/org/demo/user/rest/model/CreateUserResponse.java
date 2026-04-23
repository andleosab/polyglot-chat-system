package org.demo.user.rest.model;

import java.time.Instant;
import java.util.UUID;

public record CreateUserResponse(
        UUID userId,
        String username,
        String email,
        Instant createdAt,
        Instant updatedAt,
        Boolean isActive
) {}
