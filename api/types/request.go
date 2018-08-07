package types

import "context"

// RequestCtxKey
type requestCtxKey string

var (
	RequestIDKey = requestCtxKey("ID")
)

// RequestID creates a new request ID key for a context.
func RequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
