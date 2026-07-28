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

# Google OAuth is optional. Left empty, chat-web still starts and email/password
# sign-in works — only the "Login with Google" button fails, with the error the
# sign-in page already renders for that case. Google also requires a redirect URI
# registered up front, which the Spot node's churning NodePort IP cannot provide,
# so leave these empty until the cluster has a stable origin.
variable "google_client_id" {
  description = "Google OAuth client ID for chat-web (optional — empty disables Google sign-in)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "google_client_secret" {
  description = "Google OAuth client secret for chat-web (optional — empty disables Google sign-in)"
  type        = string
  sensitive   = true
  default     = ""
}
