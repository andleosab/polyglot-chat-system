package interceptor

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

var ConversationIDKey = contextKey{}

// ConversationIDMiddleware extracts the `id` path parameter and stores it in the request context
func ConversationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "missing conversation id", http.StatusBadRequest)
			return
		}

		conversationID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			http.Error(w, "invalid conversation_id", http.StatusBadRequest)
			return
		}

		// store it in the context
		ctx := context.WithValue(r.Context(), ConversationIDKey, conversationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper function to retrieve conversationID from context
func GetConversationID(r *http.Request) int64 {
	if val, ok := r.Context().Value(ConversationIDKey).(int64); ok {
		return val
	}
	return 0 // or handle missing value differently
}
