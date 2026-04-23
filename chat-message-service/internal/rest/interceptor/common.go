package interceptor

import "context"

// contextKey is a type to avoid context key collisions
type contextKey struct{}

func GetPrincipal(ctx context.Context) *UserPrincipal {
	p, _ := ctx.Value(principalKey).(*UserPrincipal)
	return p
}
