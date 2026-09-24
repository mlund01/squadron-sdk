package squadron

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestInvocationMetadataFromContext(t *testing.T) {
	t.Parallel()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		invocationRunIDMetadataKey, "run-1",
		invocationTaskNameMetadataKey, "collect",
		invocationAttemptIDMetadataKey, "attempt-1",
		invocationToolUseIDMetadataKey, "call-1",
		invocationIdempotencyMetadataKey, "key-1",
	))
	got, ok := InvocationMetadataFromContext(ctx)
	want := InvocationMetadata{RunID: "run-1", TaskName: "collect", AttemptID: "attempt-1", ToolUseID: "call-1", IdempotencyKey: "key-1"}
	if !ok || got != want {
		t.Fatalf("InvocationMetadataFromContext() = %#v, %v; want %#v, true", got, ok, want)
	}
}

func TestInvocationMetadataFromContextRequiresDurableIdentity(t *testing.T) {
	t.Parallel()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(invocationRunIDMetadataKey, "run-1"))
	if _, ok := InvocationMetadataFromContext(ctx); ok {
		t.Fatal("partial invocation metadata was accepted")
	}
}
