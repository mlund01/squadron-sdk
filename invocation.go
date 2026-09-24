package squadron

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	invocationRunIDMetadataKey       = "squadron-run-id"
	invocationTaskNameMetadataKey    = "squadron-task-name"
	invocationAttemptIDMetadataKey   = "squadron-attempt-id"
	invocationToolUseIDMetadataKey   = "squadron-tool-use-id"
	invocationIdempotencyMetadataKey = "squadron-idempotency-key"
)

// InvocationMetadata identifies one logical mission tool call. Squadron
// supplies it out-of-band; it is not part of the model-authored tool payload.
type InvocationMetadata struct {
	RunID          string
	TaskName       string
	AttemptID      string
	ToolUseID      string
	IdempotencyKey string
}

// InvocationMetadataFromContext exposes the durable identity attached by the
// Squadron runtime to a plugin tool call.
func InvocationMetadataFromContext(ctx context.Context) (InvocationMetadata, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return InvocationMetadata{}, false
	}
	invocation := InvocationMetadata{
		RunID:          firstMetadataValue(md, invocationRunIDMetadataKey),
		TaskName:       firstMetadataValue(md, invocationTaskNameMetadataKey),
		AttemptID:      firstMetadataValue(md, invocationAttemptIDMetadataKey),
		ToolUseID:      firstMetadataValue(md, invocationToolUseIDMetadataKey),
		IdempotencyKey: firstMetadataValue(md, invocationIdempotencyMetadataKey),
	}
	return invocation, invocation.ToolUseID != "" && invocation.IdempotencyKey != ""
}

func firstMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
