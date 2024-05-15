package otel

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

func GetContextBackground(ctx context.Context) context.Context {
	return trace.ContextWithSpan(context.Background(), trace.SpanFromContext(ctx))
}
