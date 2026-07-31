import jwt from "jsonwebtoken";
import { env } from '$env/dynamic/private';
import type { CurrentUser } from "$lib/store/user";

// Read at runtime, not at module load: the build must not require this to be set.
// symmetric HS256
function getJwtSecret(): string {
  const secret = env.JWT_SECRET;
  if (!secret) {
    throw new Error("Missing required environment variable: JWT_SECRET");
  }
  return secret;
}

// Note: Quarkus Smallrye JWT expects the secret to be in base64url format,
// so we need to encode it accordingly.
// Smallrye also is expecting byte[] secret, so we decode the base64url string to bytes before signing.
function getJwtSecretBytes(): Buffer {
  return Buffer.from(getJwtSecret().replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, ''), 'base64url');
}

export function issueServiceToken(audience: string, user: CurrentUser) {

  return jwt.sign(
    user,
    getJwtSecret(),
    {
      subject: user.userId,
      algorithm: "HS256",
      issuer: "chat-web",
      audience: audience,
      expiresIn: "10m"
    }
  );
}

export function issueWsToken(audience: string, user: CurrentUser) {

  return jwt.sign(
    {...user, scope: "ws"}, // add scope claim to distinguish from regular service tokens
    getJwtSecretBytes(),
    {
        subject: user.userId,
        algorithm: "HS256", 
        issuer: "chat-web", 
        audience: audience,
        expiresIn: "60s" }
  );
}