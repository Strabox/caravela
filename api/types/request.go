package types

import "context"

// RequestCtxKey
type requestCtxKey string

var (
	RequestIDKey = requestCtxKey("ID")
	NodeGUIDKey  = requestCtxKey("GUID")
)

// RequestID creates a new request ID key for a context.
func RequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// RequestID creates a new request ID key for a context.
func NodeGUID(ctx context.Context) string {
	if nodeGUID, ok := ctx.Value(NodeGUIDKey).(string); ok {
		return nodeGUID
	}
	return ""
}
