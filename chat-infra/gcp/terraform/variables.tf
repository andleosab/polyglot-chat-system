variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "billing_account" {
  description = "GCP billing account ID (for budget alert, format: XXXXXX-XXXXXX-XXXXXX)"
  type        = string
}

variable "jwt_secret" {
  description = "Shared HMAC-SHA256 JWT secret used by all services"
  type        = string
  sensitive   = true
}

variable "jwt_jwk" {
  description = "Base64url-encoded JWT secret required by chat-delivery-service (SmallRye JWK format)"
  type        = string
  sensitive   = true
}

variable "postgres_password" {
  description = "Password for the postgres superuser"
  type        = string
  sensitive   = true
  default     = "postgres"
}

variable "better_auth_secret" {
  description = "Secret for Better Auth session signing in chat-web"
  type        = string
  sensitive   = true
}
