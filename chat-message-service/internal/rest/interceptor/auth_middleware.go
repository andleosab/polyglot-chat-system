package interceptor

import (
	"context"
	"log"
	"net/http"
	"slices"
	"strings"

	// "github.com/go-chi/chi/v5"

	"github.com/golang-jwt/jwt/v5"

	"chat-message-service/internal/config"
	"chat-message-service/internal/rest/response"
)

type UserPrincipal struct {
	UserId   string
	Email    string
	Username string
	Roles    []string
}

var principalKey = contextKey{}

func AuthMiddleware(config *config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("==> AuthMiddleware...")

			authHeader := r.Header.Get("Authorization")
			if len(authHeader) < 7 || !strings.EqualFold(authHeader[:7], "Bearer ") {
				// http.Error(w, "Missing token", http.StatusUnauthorized)
				response.WriteJSONError(w, r, http.StatusUnauthorized, "Missing token")
				return
			}

			tokenString := authHeader[len("Bearer "):]

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
				// enforce HS256
				if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, jwt.ErrTokenSignatureInvalid
				}
				return config.Secret, nil
			},
				jwt.WithAudience(config.Audience),
				// jwt.WithExpirationRequired(),
			)

			if err != nil || !token.Valid {
				// http.Error(w, "Invalid token", http.StatusUnauthorized)
				response.WriteJSONError(w, r, http.StatusUnauthorized, "Invalid token: "+err.Error())
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				// http.Error(w, "Invalid claims", http.StatusUnauthorized)
				response.WriteJSONError(w, r, http.StatusUnauthorized, "Invalid token claims")
				return
			}

			if !validateIssuer(claims, config.Issuer) {
				response.WriteJSONError(w, r, http.StatusUnauthorized, "Invalid issuer")
				return
			}

			userID, _ := claims["sub"].(string)
			email, _ := claims["email"].(string)
			username, _ := claims["username"].(string)

			log.Println("aud:", claims["aud"])

			roles := []string{}
			if r, ok := claims["roles"].([]interface{}); ok {
				for _, v := range r {
					if s, ok := v.(string); ok {
						roles = append(roles, s)
					}
				}
			}

			principal := &UserPrincipal{
				UserId:   userID,
				Email:    email,
				Username: username,
				Roles:    roles,
			}

			ctx := context.WithValue(r.Context(), principalKey, principal)

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

func validateIssuer(claims jwt.MapClaims, allowed []string) bool {
	iss, ok := claims["iss"].(string)
	if !ok {
		return false
	}

	log.Println("issuer:", iss)
	log.Println("allowed issuers:", allowed)

	return slices.Contains(allowed, iss)
}
