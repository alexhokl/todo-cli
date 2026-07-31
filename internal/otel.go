package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// DefaultServiceName is used when OTEL_SERVICE_NAME is not set.
const DefaultServiceName = "todo"

// SetupOTel initialises the global OpenTelemetry TracerProvider, MeterProvider,
// and LoggerProvider. All three signals are exported over OTLP/gRPC to the
// collector configured through the standard OTEL_EXPORTER_OTLP_* environment
// variables, so any vendor (or a local otel-collector) can be targeted without
// code changes.
//
// Traces: a span is started for every incoming gRPC call by the otelgrpc stats
// handler registered in cmd/serve.go.
//
// Metrics: the MeterProvider is registered globally so the otelgrpc stats
// handler emits the standard gRPC server metrics automatically:
//
//   - rpc.server.duration         – latency histogram per method / status code
//   - rpc.server.request.size     – request payload size histogram
//   - rpc.server.response.size    – response payload size histogram
//   - rpc.server.requests_per_rpc – messages-per-RPC histogram
//
// Metrics are pushed at the SDK default interval (60 s), overridable via
// OTEL_METRIC_EXPORT_INTERVAL.
//
// Logs: the default slog logger is replaced with a handler backed by the
// LoggerProvider so that all slog.InfoContext / slog.WarnContext /
// slog.ErrorContext calls emit structured OTLP log records carrying the active
// trace_id and span_id, enabling log-trace correlation.
//
// When no OTLP endpoint is configured, telemetry is disabled entirely and the
// default slog handler is left untouched, so a server started without a
// collector still logs to stderr rather than silently discarding every record.
// If an endpoint is configured but cannot be used, an error is returned rather
// than being swallowed.
//
// The returned shutdown function must be deferred by the caller; it flushes and
// shuts down all three providers. It is always non-nil, even when an error is
// returned.
//
// Relevant environment variables:
//
//	OTEL_EXPORTER_OTLP_ENDPOINT  – e.g. localhost:4317; telemetry is disabled
//	                               when this and the signal-specific overrides
//	                               are all unset
//	OTEL_EXPORTER_OTLP_HEADERS   – e.g. authorization=Bearer xxx
//	OTEL_SERVICE_NAME            – e.g. todo
//	OTEL_RESOURCE_ATTRIBUTES     – additional resource attributes
//	OTEL_TRACES_SAMPLER          – e.g. parentbased_always_on (SDK default)
//	OTEL_METRIC_EXPORT_INTERVAL  – metric push interval in ms (default 60000)
func SetupOTel(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls each registered cleanup function and joins their errors.
	shutdown = func(ctx context.Context) error {
		var errs []error
		for _, fn := range shutdownFuncs {
			errs = append(errs, fn(ctx))
		}
		return errors.Join(errs...)
	}

	// Without a configured collector there is nowhere to export to. Returning
	// early leaves the default slog handler in place so log records continue to
	// reach stderr instead of being dropped by the OTLP bridge.
	if !isOTLPConfigured() {
		slog.Debug("OpenTelemetry is disabled as no OTLP endpoint is configured")
		return shutdown, nil
	}

	// Build a shared resource describing this service.
	res, err := buildResource(ctx)
	if err != nil {
		return shutdown, fmt.Errorf("failed to build OpenTelemetry resource: %w", err)
	}

	// ── Traces ────────────────────────────────────────────────────────────────

	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return shutdown, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	shutdownFuncs = append(shutdownFuncs, tp.Shutdown)

	// Register as the global TracerProvider so otelgrpc (and any other
	// instrumentation) picks it up automatically.
	otel.SetTracerProvider(tp)

	// W3C TraceContext + Baggage propagation, required for cross-service
	// correlation.
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// ── Logs ──────────────────────────────────────────────────────────────────

	logExporter, err := otlploggrpc.New(ctx)
	if err != nil {
		return shutdown, fmt.Errorf("failed to create log exporter: %w", err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)
	shutdownFuncs = append(shutdownFuncs, lp.Shutdown)

	// Replace the default slog logger. The otelslog bridge handler reads the
	// active span from the context passed to slog.XxxContext calls and
	// populates the OTLP LogRecord's trace_id, span_id, and trace_flags fields
	// automatically — no custom handler code is required.
	handler := otelslog.NewHandler(serviceName(), otelslog.WithLoggerProvider(lp))
	slog.SetDefault(slog.New(handler))

	// ── Metrics ───────────────────────────────────────────────────────────────

	metricExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return shutdown, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		// PeriodicReader with no explicit interval: the SDK default of 60 s is
		// used, overridable at runtime via OTEL_METRIC_EXPORT_INTERVAL (ms).
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	shutdownFuncs = append(shutdownFuncs, mp.Shutdown)

	// Register as the global MeterProvider. The otelgrpc stats handler in
	// cmd/serve.go picks this up automatically.
	otel.SetMeterProvider(mp)

	// Route OTel SDK internal errors (including failed OTLP exports due to auth
	// rejection, TLS errors, or timeouts) to stderr so they are visible in the
	// server logs. Without this, batch processor export failures are silently
	// discarded and there is no indication that telemetry is not being
	// delivered.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		fmt.Fprintf(os.Stderr, "OpenTelemetry error: %v\n", err)
	}))

	return shutdown, nil
}

// isOTLPConfigured reports whether an OTLP endpoint has been configured through
// the standard environment variables, either globally or for any individual
// signal.
func isOTLPConfigured() bool {
	for _, name := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
	} {
		if os.Getenv(name) != "" {
			return true
		}
	}
	return false
}

// serviceName resolves the service name from the environment, falling back to
// DefaultServiceName.
func serviceName() string {
	if name := os.Getenv("OTEL_SERVICE_NAME"); name != "" {
		return name
	}
	return DefaultServiceName
}

func buildResource(ctx context.Context) (*resource.Resource, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return resource.New(ctx,
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithFromEnv(), // picks up OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME
		resource.WithAttributes(
			semconv.ServiceName(serviceName()),
			semconv.ServiceInstanceID(hostname),
			attribute.String("host.name", hostname),
		),
	)
}
