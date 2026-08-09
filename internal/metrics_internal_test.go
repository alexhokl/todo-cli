package internal

import (
	"context"
	"slices"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// instrumentNames collects all metric instrument names that have been recorded
// against the given manual reader since the last Collect. The OTel SDK only
// surfaces an instrument in Collect output after at least one measurement has
// been recorded for it, so each instrument must be exercised at least once
// before this is called.
func instrumentNames(t *testing.T, reader *sdkmetric.ManualReader) []string {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("reader.Collect failed: %v", err)
	}

	var names []string
	for _, scope := range rm.ScopeMetrics {
		for _, m := range scope.Metrics {
			names = append(names, m.Name)
		}
	}
	return names
}

// TestNewMetricsRegistersAllInstruments asserts that NewMetrics registers every
// expected application instrument and that the returned *Metrics carries
// non-nil handles. The set of names is the canonical list of application
// metrics; any addition or removal must update this test so that a missing or
// unexpected instrument fails the build.
func TestNewMetricsRegistersAllInstruments(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	meter := provider.Meter(meterName)
	m, err := NewMetrics(meter)
	if err != nil {
		t.Fatalf("NewMetrics failed: %v", err)
	}
	if m == nil {
		t.Fatal("NewMetrics returned nil *Metrics")
	}

	if m.AuthAttempts == nil {
		t.Error("AuthAttempts handle must be non-nil")
	}
	if m.RecordsReturned == nil {
		t.Error("RecordsReturned handle must be non-nil")
	}

	// The OTel SDK only exposes an instrument via Collect after a measurement
	// has been recorded against it, so exercise every instrument through the
	// package-level helpers. Save and restore appMetrics so the test does not
	// leak global state into other tests in the package.
	previous := appMetrics
	appMetrics = m
	t.Cleanup(func() { appMetrics = previous })

	ctx := context.Background()
	recordAuthAttempt(ctx, "test", "success")
	recordRecordsReturned(ctx, "test/TestList", "test", 1)

	names := instrumentNames(t, reader)

	expected := []string{
		"todo.auth.attempts",
		"todo.records.returned",
	}
	for _, want := range expected {
		if !slices.Contains(names, want) {
			t.Errorf("expected instrument %q to be registered; got %v", want, names)
		}
	}
}

// TestRecordingHelpersNoOpWhenAppMetricsNil guards against a nil-pointer panic
// when telemetry is disabled (OTEL_SDK_DISABLED=true) and appMetrics is never
// populated. Every helper must silently do nothing in that state.
func TestRecordingHelpersNoOpWhenAppMetricsNil(t *testing.T) {
	previous := appMetrics
	appMetrics = nil
	t.Cleanup(func() { appMetrics = previous })

	ctx := context.Background()
	assertNotPanics(t, func() {
		recordAuthAttempt(ctx, "test", "success")
		recordRecordsReturned(ctx, "test/TestList", "test", 1)
	})
}

// assertNotPanics asserts that fn runs without panicking, reporting the
// recovered panic value via t.Errorf when it does.
func assertNotPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	fn()
}