package org.demo.user.service;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

import org.demo.user.rest.model.CreateUserRequest;
import org.demo.user.rest.model.CreateUserResponse;

public interface UserService {
	
	CreateUserResponse createUser(CreateUserRequest request);
	Optional<CreateUserResponse> getUserByUserId(UUID userId);
	List<CreateUserResponse> getAllUsers();
	void deactivateUser(UUID userId);

}
