resource "kubernetes_namespace_v1" "chat" {
  metadata {
    name = "chat"
  }
  depends_on = [google_container_node_pool.main]
}

# Shared secrets referenced by all service deployments
resource "kubernetes_secret_v1" "chat_secrets" {
  metadata {
    name      = "chat-secrets"
    namespace = kubernetes_namespace_v1.chat.metadata[0].name
  }

  data = {
    jwt-secret         = var.jwt_secret
    jwt-jwk            = var.jwt_jwk
    postgres-password  = var.postgres_password
    better-auth-secret = var.better_auth_secret

    # Optional — empty by default, which leaves Google sign-in disabled. The keys are
    # always present so chat-web's secretKeyRef resolves either way.
    google-client-id     = var.google_client_id
    google-client-secret = var.google_client_secret

    # Optional — empty until a Cloudflare tunnel exists. deploy-infra-services.sh
    # reads this key to decide whether to deploy cloudflared.
    tunnel-token = var.tunnel_token

    # Full connection strings per service — avoids plaintext passwords in values files
    db-url-auth     = "postgresql://postgres:${var.postgres_password}@postgres.chat.svc.cluster.local:5432/chat-auth"
    db-url-user     = "postgresql://postgres:${var.postgres_password}@postgres.chat.svc.cluster.local:5432/chat-user"
    db-url-message  = "postgres://postgres:${var.postgres_password}@postgres.chat.svc.cluster.local:5432/chat-message?sslmode=disable"
    db-url-presence = "postgresql://postgres:${var.postgres_password}@postgres.chat.svc.cluster.local:5432/chat-presence"
  }
}
