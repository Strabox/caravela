package types

import "context"

// RequestCtxKey
type requestCtxKey string

var (
	RequestIDKey       = requestCtxKey("ID")
	NodeGUIDKey        = requestCtxKey("GUID")
	PartitionsStateKey = requestCtxKey("PartitionsState")
)

// RequestID retrieves the request ID key from a context.
func RequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// NodeGUID retrieves the node's GUID key from a context.
func NodeGUID(ctx context.Context) string {
	if nodeGUID, ok := ctx.Value(NodeGUIDKey).(string); ok {
		return nodeGUID
	}
	return ""
}

// SysPartitionsState retrieves the partitions state from a context.
func SysPartitionsState(ctx context.Context) []PartitionState {
	if partitionsState, ok := ctx.Value(PartitionsStateKey).([]PartitionState); ok {
		return partitionsState
	}
	return nil
}
