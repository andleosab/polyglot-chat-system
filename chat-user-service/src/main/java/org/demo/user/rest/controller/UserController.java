package org.demo.user.rest.controller;

import java.util.List;
import java.util.UUID;

import org.demo.user.rest.model.CreateUserRequest;
import org.demo.user.rest.model.CreateUserResponse;
import org.demo.user.service.UserService;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;

@RequiredArgsConstructor
@RestController
@RequestMapping("/api/users")
public class UserController {
	
	private final UserService userService;
	
	@PostMapping(value = "", consumes = MediaType.APPLICATION_JSON_VALUE, 
			produces = MediaType.APPLICATION_JSON_VALUE)
	public ResponseEntity<CreateUserResponse> createUser(@Valid @RequestBody CreateUserRequest request) {
    	
        return ResponseEntity.ok(userService.createUser(request));
    }	

    @GetMapping("/{id}")
    public ResponseEntity<CreateUserResponse> getUser(@PathVariable UUID id) {
        return userService.getUserByUserId(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }
    
    @GetMapping
    public List<CreateUserResponse> getAllUsers() {
        return userService.getAllUsers();
    }

    @PutMapping("/{id}/deactivate")
    public ResponseEntity<Void> deactivate(@PathVariable UUID id) {
        userService.deactivateUser(id);
        return ResponseEntity.noContent().build();
    }    
}
