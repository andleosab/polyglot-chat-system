package interceptor

import (
	"net/http"
	"slices"
)

// RequireRoleMiddleware checks if the authenticated user has the required role to access the endpoint.
// It should be used after the AuthMiddleware, which populates the user principal in the context.
// Usage:
// Per route:
// r.With(RequireRoleMiddleware("admin")).Get("/admin", adminHandler)
// Or for a group of routes:
//
//	r.Route("/admin", func(r chi.Router) {
//	    r.Use(RequireRoleMiddleware("admin"))
//	    r.Get("/", adminHandler)
//	    r.Get("/settings", adminSettingsHandler)
//	})
func RequireRoleMiddleware(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			user := GetPrincipal(r.Context())
			if user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if slices.Contains(user.Roles, role) {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}
