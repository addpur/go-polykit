package telemetry

import (
	"context"

	"github.com/addpur/go-polykit"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware is a stub OpenTelemetry middleware that starts and ends a span.
func TracingMiddleware(endpointName string) polykit.Middleware {
	return func(next polykit.Endpoint) polykit.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			tracer := otel.Tracer("go-polykit/telemetry")
			
			ctx, span := tracer.Start(ctx, endpointName, trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			// In a real implementation, you might want to record errors or other attributes on the span here.
			// e.g., span.RecordError(err)

			return next(ctx, request)
		}
	}
}
