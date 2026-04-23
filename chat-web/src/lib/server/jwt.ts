import jwt from "jsonwebtoken";
import { env } from '$env/dynamic/private';
import type { CurrentUser } from "$lib/store/user";

// symmetric HS256 
const JWT_SECRET = env.JWT_SECRET!;

// Note: Quarkus Smallrye JWT expects the secret to be in base64url format, 
// so we need to encode it accordingly.
// Smallrye also is expecting byte[] secret, so we decode the base64url string to bytes before signing. 
const JWT_SECRET_BYTES = Buffer.from(JWT_SECRET.replace(/\+/g, '-')
        .replace(/\//g, '_').replace(/=+$/, ''), 'base64url');

export function issueServiceToken(audience: string, user: CurrentUser) {

  return jwt.sign(
    user,
    JWT_SECRET,
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
    JWT_SECRET_BYTES,
    {
        subject: user.userId,
        algorithm: "HS256", 
        issuer: "chat-web", 
        audience: audience,
        expiresIn: "60s" }
  );
}