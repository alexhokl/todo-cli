package cmd

import (
	"testing"

	"github.com/alexhokl/todo-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestValidateFlags(t *testing.T) {
	tests := []struct {
		name        string
		opts        serveOptions
		expectError bool
	}{
		{"valid", serveOptions{Port: 8080, DatabaseFilePath: "/tmp/todo.db"}, false},
		{"valid with hostname", serveOptions{Port: 443, DatabaseFilePath: "/tmp/todo.db", Hostname: "todo", TailscaleAuthKey: "tskey"}, false},
		{"lowest valid port", serveOptions{Port: 1, DatabaseFilePath: "/tmp/todo.db"}, false},
		{"highest valid port", serveOptions{Port: 65535, DatabaseFilePath: "/tmp/todo.db"}, false},
		{"zero port", serveOptions{Port: 0, DatabaseFilePath: "/tmp/todo.db"}, true},
		{"negative port", serveOptions{Port: -1, DatabaseFilePath: "/tmp/todo.db"}, true},
		{"port out of range", serveOptions{Port: 65536, DatabaseFilePath: "/tmp/todo.db"}, true},
		{"empty database path", serveOptions{Port: 8080, DatabaseFilePath: ""}, true},
		{"hostname without auth key", serveOptions{Port: 8080, DatabaseFilePath: "/tmp/todo.db", Hostname: "todo"}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateFlags(test.opts)
			if test.expectError && err == nil {
				t.Errorf("expected an error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
		})
	}
}

func TestBindEnvironmentVariablesToServeOptions(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		viperValues      map[string]any
		expectedPort     int
		expectedDatabase string
	}{
		{
			name:             "flag takes precedence over viper",
			args:             []string{"--port", "9090", "--database", "/from/flag.db"},
			viperValues:      map[string]any{"port": 7070, "database": "/from/viper.db"},
			expectedPort:     9090,
			expectedDatabase: "/from/flag.db",
		},
		{
			name:             "viper fills in unchanged flags",
			args:             []string{},
			viperValues:      map[string]any{"port": 7070, "database": "/from/viper.db"},
			expectedPort:     7070,
			expectedDatabase: "/from/viper.db",
		},
		{
			name:             "defaults retained when viper is empty",
			args:             []string{},
			viperValues:      map[string]any{},
			expectedPort:     DefaultPort,
			expectedDatabase: "/default.db",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			for key, value := range test.viperValues {
				viper.Set(key, value)
			}

			opts := serveOptions{Port: DefaultPort, DatabaseFilePath: "/default.db"}
			cmd := &cobra.Command{Use: "serve", RunE: func(*cobra.Command, []string) error { return nil }}
			flags := cmd.Flags()
			flags.IntVarP(&opts.Port, "port", "p", DefaultPort, "")
			flags.StringVarP(&opts.DatabaseFilePath, "database", "d", "/default.db", "")
			flags.StringVar(&opts.Hostname, "hostname", "", "")
			flags.StringVar(&opts.TailscaleAuthKey, "ts-auth-key", "", "")
			flags.StringVar(&opts.TailscaleStateDirectory, "ts-state-dir", "./tailscale-state", "")

			cmd.SetArgs(test.args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("failed to parse flags: %v", err)
			}

			bindEnvironmentVariablesToServeOptions(cmd, &opts)

			if opts.Port != test.expectedPort {
				t.Errorf("expected port %d but got %d", test.expectedPort, opts.Port)
			}
			if opts.DatabaseFilePath != test.expectedDatabase {
				t.Errorf("expected database %q but got %q", test.expectedDatabase, opts.DatabaseFilePath)
			}
		})
	}
}

func TestBindEnvironmentVariablesAppliesTailscaleOptions(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		viperValues    map[string]any
		expectedHost   string
		expectedKey    string
		expectedState  string
	}{
		{
			name:         "defaults when nothing set",
			args:         []string{},
			viperValues:  map[string]any{},
			expectedHost: "",
			expectedKey:  "",
			expectedState: "./tailscale-state",
		},
		{
			name:         "flag takes precedence over viper",
			args:         []string{"--hostname", "todo", "--ts-auth-key", "tskey-flag", "--ts-state-dir", "/flag-state"},
			viperValues:  map[string]any{"hostname": "viper", "ts_auth_key": "tskey-viper", "ts_state_dir": "/viper-state"},
			expectedHost: "todo",
			expectedKey:  "tskey-flag",
			expectedState: "/flag-state",
		},
		{
			name:         "viper fills in unchanged flags",
			args:         []string{},
			viperValues:  map[string]any{"hostname": "todo", "ts_auth_key": "tskey-viper", "ts_state_dir": "/viper-state"},
			expectedHost: "todo",
			expectedKey:  "tskey-viper",
			expectedState: "/viper-state",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			for key, value := range test.viperValues {
				viper.Set(key, value)
			}

			opts := serveOptions{TailscaleStateDirectory: "./tailscale-state"}
			cmd := &cobra.Command{Use: "serve", RunE: func(*cobra.Command, []string) error { return nil }}
			flags := cmd.Flags()
			flags.StringVar(&opts.Hostname, "hostname", "", "")
			flags.StringVar(&opts.TailscaleAuthKey, "ts-auth-key", "", "")
			flags.StringVar(&opts.TailscaleStateDirectory, "ts-state-dir", "./tailscale-state", "")

			cmd.SetArgs(test.args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("failed to parse flags: %v", err)
			}

			bindEnvironmentVariablesToServeOptions(cmd, &opts)

			if opts.Hostname != test.expectedHost {
				t.Errorf("expected hostname %q but got %q", test.expectedHost, opts.Hostname)
			}
			if opts.TailscaleAuthKey != test.expectedKey {
				t.Errorf("expected ts-auth-key %q but got %q", test.expectedKey, opts.TailscaleAuthKey)
			}
			if opts.TailscaleStateDirectory != test.expectedState {
				t.Errorf("expected ts-state-dir %q but got %q", test.expectedState, opts.TailscaleStateDirectory)
			}
		})
	}
}

func TestGetGrpcServer(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	server := getGrpcServer(db, internal.DummyAuthenticationInterceptor, internal.DummyStreamAuthenticationInterceptor)
	if server == nil {
		t.Fatalf("expected a gRPC server but got nil")
	}
	defer server.Stop()

	// Reflection is registered so grpcurl works before any service exists.
	if _, ok := server.GetServiceInfo()["grpc.reflection.v1.ServerReflection"]; !ok {
		t.Errorf("expected server reflection to be registered, got %v", server.GetServiceInfo())
	}

	if _, ok := server.GetServiceInfo()["todo.TodoService"]; !ok {
		t.Errorf("expected the todo service to be registered, got %v", server.GetServiceInfo())
	}
}

func TestDefaultDatabaseFilePath(t *testing.T) {
	if path := defaultDatabaseFilePath(); path == "" {
		t.Errorf("expected a non-empty default database file path")
	}
}