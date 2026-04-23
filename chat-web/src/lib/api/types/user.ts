/**
 * Based on OpenAPI definition 1.0.0-oas3.1
 * Matches your Chat System's User Management requirements.
 */

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  /** @format uuid */
  userId: string;
  username: string;
  email: string;
  isActive: boolean;
}

export interface CreateUserRequest {
  /** @minLength 3 @maxLength 50 */
  username: string;
  /** @format email @minLength 1 */
  email: string;
  /** @format uuid */
  userid: string; 
}

export interface CreateUserResponse {
  /** @format uuid */
  userId: string;
  username: string;
  email: string;
  /** @format date-time */
  createdAt: string;
  /** @format date-time */
  updatedAt: string;
  isActive: boolean;
}

/**
 * Helper type for API Path parameters
 */
export type UserIdParam = string; // UUID string