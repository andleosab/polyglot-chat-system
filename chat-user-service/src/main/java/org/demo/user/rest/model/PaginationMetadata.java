package org.demo.user.rest.model;

public record PaginationMetadata(
	    int page,
	    int pageSize,
	    int count,
	    long total,
	    int totalPages,
	    boolean hasNext,
	    boolean hasPrev
) {}

