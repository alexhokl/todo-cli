package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestBindEnvironmentVariablesToRootOptions(t *testing.T) {
	tests := []struct {
		name             string
		initial          rootOptions
		viperValues      map[string]any
		expectedService  string
		expectedInsecure bool
	}{
		{
			name:             "flag takes precedence over viper",
			initial:          rootOptions{serviceURI: "from-flag"},
			viperValues:      map[string]any{"service": "from-viper"},
			expectedService:  "from-flag",
			expectedInsecure: false,
		},
		{
			name:             "viper fills in unset flag",
			initial:          rootOptions{},
			viperValues:      map[string]any{"service": "from-viper", "insecure": true},
			expectedService:  "from-viper",
			expectedInsecure: true,
		},
		{
			name:             "no value anywhere",
			initial:          rootOptions{},
			viperValues:      map[string]any{},
			expectedService:  "",
			expectedInsecure: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			for key, value := range test.viperValues {
				viper.Set(key, value)
			}

			opts := test.initial
			bindEnvironmentVariablesToRootOptions(&opts)

			if opts.serviceURI != test.expectedService {
				t.Errorf("expected service %q but got %q", test.expectedService, opts.serviceURI)
			}
			if opts.allowInsecureConnection != test.expectedInsecure {
				t.Errorf("expected insecure %v but got %v", test.expectedInsecure, opts.allowInsecureConnection)
			}
		})
	}
}

func TestRequiresService(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *cobra.Command
		expected bool
	}{
		{"nil command", nil, false},
		{"no annotations", &cobra.Command{Use: "list"}, false},
		{
			"annotated false",
			&cobra.Command{Use: "serve", Annotations: map[string]string{annotationRequiresService: "false"}},
			false,
		},
		{
			"annotated true",
			&cobra.Command{Use: "get", Annotations: map[string]string{annotationRequiresService: "true"}},
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if result := requiresService(test.cmd); result != test.expected {
				t.Errorf("expected %v but got %v", test.expected, result)
			}
		})
	}
}

func TestServeCommandDoesNotRequireService(t *testing.T) {
	if requiresService(serveCmd) {
		t.Errorf("serve command must not require a service URI")
	}
}
