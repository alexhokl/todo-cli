package internal

import (
	"context"
	"log/slog"
	"testing"

	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func TestIsOTLPConfigured(t *testing.T) {
	envNames := []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
	}

	tests := []struct {
		name     string
		set      string
		expected bool
	}{
		{"none set", "", false},
		{"general endpoint", "OTEL_EXPORTER_OTLP_ENDPOINT", true},
		{"traces endpoint", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", true},
		{"metrics endpoint", "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", true},
		{"logs endpoint", "OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, name := range envNames {
				t.Setenv(name, "")
			}
			if test.set != "" {
				t.Setenv(test.set, "localhost:4317")
			}
			if result := isOTLPConfigured(); result != test.expected {
				t.Errorf("expected %v but got %v", test.expected, result)
			}
		})
	}
}

// TestSetupOTelWithoutEndpoint asserts that a server started without a
// collector keeps its default slog handler, so log records are not silently
// discarded.
func TestSetupOTelWithoutEndpoint(t *testing.T) {
	for _, name := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
	} {
		t.Setenv(name, "")
	}

	before := slog.Default()

	shutdown, err := SetupOTel(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shutdown == nil {
		t.Fatalf("expected a non-nil shutdown function")
	}
	if slog.Default() != before {
		t.Errorf("expected the default slog logger to be left untouched")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("expected no error from shutdown, got %v", err)
	}
}

func TestServiceName(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{"unset", "", DefaultServiceName},
		{"set", "custom-todo", "custom-todo"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("OTEL_SERVICE_NAME", test.envValue)
			if result := serviceName(); result != test.expected {
				t.Errorf("expected %v but got %v", test.expected, result)
			}
		})
	}
}

func TestBuildResource(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "todo-test")

	res, err := buildResource(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var found bool
	for _, attr := range res.Attributes() {
		if attr.Key == semconv.ServiceNameKey && attr.Value.AsString() == "todo-test" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected resource to carry service.name of todo-test")
	}
}
