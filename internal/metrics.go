package internal

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const meterName = "todo"

// appMetrics is the package-level handle to the registered application
// metrics. It is populated by SetupOTel once the global MeterProvider has been
// installed, and read by the gRPC handlers and the authentication interceptors
// when they record business-level measurements. It remains nil when telemetry
// is disabled (OTEL_SDK_DISABLED=true) or when instrument registration fails;
// every recording site must guard against a nil value so a missing meter never
// panics the request path.
var appMetrics *Metrics

// Metrics holds the OpenTelemetry instruments used to record
// application-level measurements for the todo gRPC API. The instruments
// complement the automatic rpc.server.* metrics emitted by the otelgrpc
// stats handler with business dimensions the stats handler cannot see: record
// counts returned by List RPCs and authentication attempt outcomes.
//
// All instruments are registered once at startup via NewMetrics and the
// returned *Metrics is stored in the package-level appMetrics variable by
// SetupOTel. Recording sites use the small helper methods below, each of
// which is a no-op when appMetrics is nil so that handlers do not need to
// repeat the nil check at every call site.
type Metrics struct {
	// AuthAttempts counts every authentication attempt observed by the
	// authentication interceptors, labelled by the auth mode (tailscale or
	// dummy) and the outcome (success or failure).
	AuthAttempts metric.Int64Counter

	// RecordsReturned records the number of records returned by each List
	// RPC, labelled by the gRPC method name and the entity kind. It is a
	// histogram rather than a counter so that downstream dashboards can
	// compute p50/p95 page sizes as well as totals.
	RecordsReturned metric.Int64Histogram
}

// Attribute keys used by the application metrics. Centralised here so the
// spelling is consistent across recording sites.
const (
	attrAuthMode   = attribute.Key("auth.mode")
	attrAuthResult = attribute.Key("auth.result")
	attrRPCMethod  = attribute.Key("rpc.method")
	attrEntity     = attribute.Key("entity")
)

// Explicit bucket boundaries for the histograms. The OTel SDK default
// boundaries are tuned for latency in seconds and produce useless buckets for
// record counts, so the RecordsReturned instrument overrides them.
var recordsReturnedBuckets = []float64{1, 10, 50, 100, 500, 1000}

// NewMetrics registers all application-level instruments on the provided meter
// and returns a Metrics handle holding them. The meter is expected to be the
// global one (otel.Meter(meterName)) so that the PeriodicReader in SetupOTel
// exports the recorded data points; passing a separate meter would silently
// orphan the instruments.
//
// Registration errors are returned to the caller, which is expected to log
// them and continue (matching the non-fatal telemetry pattern in SetupOTel).
func NewMetrics(meter metric.Meter) (*Metrics, error) {
	authAttempts, err := meter.Int64Counter(
		"todo.auth.attempts",
		metric.WithDescription("Number of authentication attempts observed by the auth interceptors"),
	)
	if err != nil {
		return nil, err
	}

	recordsReturned, err := meter.Int64Histogram(
		"todo.records.returned",
		metric.WithDescription("Number of records returned by List RPCs"),
		metric.WithExplicitBucketBoundaries(recordsReturnedBuckets...),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		AuthAttempts:    authAttempts,
		RecordsReturned: recordsReturned,
	}, nil
}

// --- Recording helpers -------------------------------------------------
//
// Each helper is a no-op when appMetrics is nil. The nil check is deliberately
// kept inside the helper rather than at every call site so that handlers stay
// focused on business logic; a missing meter (because telemetry is disabled or
// registration failed) must never panic the request path.

// recordAuthAttempt increments the auth.attempts counter with the given mode
// and result labels. The context is used by the OTel SDK for baggage and
// active-span correlation.
func recordAuthAttempt(ctx context.Context, mode, result string) {
	if appMetrics == nil {
		return
	}
	appMetrics.AuthAttempts.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrAuthMode.String(mode),
			attrAuthResult.String(result),
		),
	)
}

// recordRecordsReturned records the number of records returned by a List RPC,
// labelled by the gRPC method name and the entity kind.
func recordRecordsReturned(ctx context.Context, method, entity string, count int) {
	if appMetrics == nil {
		return
	}
	appMetrics.RecordsReturned.Record(
		ctx,
		int64(count),
		metric.WithAttributes(
			attrRPCMethod.String(method),
			attrEntity.String(entity),
		),
	)
}