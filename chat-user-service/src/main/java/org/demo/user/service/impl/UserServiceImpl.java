package org.demo.user.service.impl;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

import org.demo.user.data.entity.UserEntity;
import org.demo.user.data.repository.UserRepository;
import org.demo.user.mapper.UserMapper;
import org.demo.user.rest.model.CreateUserRequest;
import org.demo.user.rest.model.CreateUserResponse;
import org.demo.user.service.UserService;
import org.springframework.stereotype.Service;

import jakarta.transaction.Transactional;
import lombok.RequiredArgsConstructor;

@RequiredArgsConstructor
@Service
@Transactional
public class UserServiceImpl implements UserService {
	
	private final UserRepository userRepository;

	@Override
	public CreateUserResponse createUser(CreateUserRequest request) {
        UserEntity user = new UserEntity();
        user.setUsername(request.username());
        user.setEmail(request.email());
        user.setUserId(request.userid());
        UserEntity saved = userRepository.save(user);
        return UserMapper.toCreateUserResponse(saved);
	}

	@Override
	public Optional<CreateUserResponse> getUserByUserId(UUID userId) {
        return userRepository.findByUserId(userId)
                .map(UserMapper::toCreateUserResponse);
	}

	@Override
	public List<CreateUserResponse> getAllUsers() {
        return userRepository.findAll()
                .stream()
                .map(UserMapper::toCreateUserResponse)
                .toList();
	}

	@Override
	public void deactivateUser(UUID userId) {
        userRepository.findByUserId(userId).ifPresent(user -> {
            user.setIsActive(false);
            userRepository.save(user);
        });
	}

}
