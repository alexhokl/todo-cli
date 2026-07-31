package cmd

import "testing"

func TestRequireSecureConnection(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"", false},
		{"localhost", false},
		{"localhost:8080", false},
		{"127.0.0.1:8080", false},
		{"[::1]:8080", false},
		{"todo.example.com", true},
		{"todo.example.com:443", true},
	}

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			result := requireSecureConnection(test.url)
			if result != test.expected {
				t.Errorf("expected %v but got %v", test.expected, result)
			}
		})
	}
}

func TestGetConnectionCredentials(t *testing.T) {
	tests := []struct {
		name     string
		secure   bool
		expected string
	}{
		{"insecure", false, "insecure"},
		{"secure", true, "tls"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			creds := getConnectionCredentials(test.secure)
			if creds == nil {
				t.Fatalf("expected credentials but got nil")
			}
			if protocol := creds.Info().SecurityProtocol; protocol != test.expected {
				t.Errorf("expected security protocol %q but got %q", test.expected, protocol)
			}
		})
	}
}
