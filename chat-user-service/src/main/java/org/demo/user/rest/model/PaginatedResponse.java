package org.demo.user.rest.model;

import java.util.List;

public record PaginatedResponse<T>(
    List<T> data,
    PaginationMetadata pagination		
) {}
