variable "project_id" {
  description = "GCP project ID"
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

# Empty until a Cloudflare tunnel exists. deploy-infra-services.sh skips the
# cloudflared release while this is blank, which leaves kubectl port-forward as the
# only way into the cluster — the intended state before a domain is in place.
variable "tunnel_token" {
  description = "Cloudflare Tunnel token for cloudflared (optional — empty leaves the tunnel undeployed)"
  type        = string
  sensitive   = true
  default     = ""
}

# Google OAuth is optional. Left empty, chat-web still starts and email/password
# sign-in works — only the "Login with Google" button fails, with the error the
# sign-in page already renders for that case. Google also requires a redirect URI
# registered up front, which needs a stable origin — the tunnel hostname provides
# one, so these become usable once the tunnel is live.
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
