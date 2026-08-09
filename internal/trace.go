package internal

import (
	"context"

	"github.com/alexhokl/helper/telemetry"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "todo"

// startSpan starts a new child span with the given name using the package tracer.
func startSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return telemetry.StartSpan(ctx, tracerName, spanName)
}

// endSpanOk marks the span status as Ok and ends it.
func endSpanOk(span trace.Span) {
	telemetry.EndSpanOk(span)
}
