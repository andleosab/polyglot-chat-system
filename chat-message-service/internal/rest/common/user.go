package common

import "net/http"

// User type for CurrentUser context
type User struct {
	ID    string
	Email string
	Roles []string
}

// context key
type CtxKey string

const UserCtxKey CtxKey = "currentUser"

// helper to retrieve user from request context
func CurrentUser(r *http.Request) *User {
	user, ok := r.Context().Value(UserCtxKey).(*User)
	if !ok {
		return nil
	}
	return user
}
