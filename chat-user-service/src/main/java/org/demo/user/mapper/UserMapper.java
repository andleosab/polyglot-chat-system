package org.demo.user.mapper;

import org.demo.user.data.entity.UserEntity;
import org.demo.user.rest.model.CreateUserResponse;

public class UserMapper {

    public static CreateUserResponse toCreateUserResponse(UserEntity entity) {
        return new CreateUserResponse(
                entity.getUserId(),
                entity.getUsername(),
                entity.getEmail(),
                entity.getCreatedAt(),
                entity.getUpdatedAt(),
                entity.getIsActive()
        );
    }
    
}
