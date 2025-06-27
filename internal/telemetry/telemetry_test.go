package telemetry

import (
	"context"
	"testing"
	"time"

	"gemini-cli-go/internal/config"
)

func TestInitializeTelemetryDisabled(t *testing.T) {
	// Reset the global tp for this test
	originalTp := tp
	tp = nil
	defer func() { tp = originalTp }()
	
	// Test with telemetry disabled
	cfg := &config.CliConfig{
		TelemetryEnabled: boolPtr(false),
	}

	// This should not panic or error
	InitializeTelemetry(cfg)

	// Verify that tp is still nil (not initialized)
	if tp != nil {
		t.Error("Expected telemetry provider to be nil when disabled")
	}
}

func TestInitializeTelemetryNilConfig(t *testing.T) {
	// Reset the global tp for this test
	originalTp := tp
	tp = nil
	defer func() { tp = originalTp }()
	
	// Test with nil telemetry enabled (should default to disabled)
	cfg := &config.CliConfig{
		TelemetryEnabled: nil,
	}

	// This should not panic or error
	InitializeTelemetry(cfg)

	// Verify that tp is still nil (not initialized)
	if tp != nil {
		t.Error("Expected telemetry provider to be nil when TelemetryEnabled is nil")
	}
}

func TestShutdownTelemetryNotInitialized(t *testing.T) {
	// Ensure tp is nil
	originalTp := tp
	tp = nil
	defer func() { tp = originalTp }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This should not panic or error
	ShutdownTelemetry(ctx)
}

func TestInitializeTelemetryEnabled(t *testing.T) {
	// Skip this test in most cases as it requires an OTLP endpoint
	// This is more of an integration test
	if testing.Short() {
		t.Skip("Skipping telemetry integration test in short mode")
	}

	// Test with telemetry enabled but invalid endpoint
	// This will fail to connect but should not panic
	cfg := &config.CliConfig{
		TelemetryEnabled:      boolPtr(true),
		TelemetryOtlpEndpoint: "http://localhost:4317", // Likely non-existent endpoint
	}

	// This should not panic, even if it fails to connect
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("InitializeTelemetry panicked: %v", r)
		}
	}()

	// Note: This will likely fail to connect to the OTLP endpoint,
	// but it should not panic. In a real scenario, the endpoint would be valid.
	// For unit testing, we're mainly testing that the function doesn't crash.
	InitializeTelemetry(cfg)

	// If we get here without panicking, the test passes
	// In a real scenario with a valid endpoint, tp would be set
}

func TestShutdownTelemetryWithProvider(t *testing.T) {
	// Skip this test in most cases as it requires initialization
	if testing.Short() {
		t.Skip("Skipping telemetry shutdown test in short mode")
	}

	// This test would require a properly initialized provider
	// For now, we'll test the case where tp is not nil but shutdown might fail
	// This is more of an integration test and would need a real OTLP setup

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// If tp was initialized in a previous test, try to shut it down
	if tp != nil {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShutdownTelemetry panicked: %v", r)
			}
		}()

		ShutdownTelemetry(ctx)
	}
}

func TestTelemetryLifecycle(t *testing.T) {
	// Test the complete lifecycle with disabled telemetry
	// Reset the global tp for this test
	originalTp := tp
	tp = nil
	defer func() { tp = originalTp }()
	
	cfg := &config.CliConfig{
		TelemetryEnabled: boolPtr(false),
	}

	// Initialize
	InitializeTelemetry(cfg)
	
	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ShutdownTelemetry(ctx)

	// Verify no issues
	if tp != nil {
		t.Error("Expected telemetry provider to remain nil throughout lifecycle when disabled")
	}
}

func TestTelemetryConfigValidation(t *testing.T) {
	// Test various configuration scenarios
	testCases := []struct {
		name           string
		cfg            *config.CliConfig
		expectInit     bool
	}{
		{
			name: "Enabled with endpoint",
			cfg: &config.CliConfig{
				TelemetryEnabled:      boolPtr(true),
				TelemetryOtlpEndpoint: "http://localhost:4317",
			},
			expectInit: true, // Would init if endpoint was valid
		},
		{
			name: "Enabled without endpoint",
			cfg: &config.CliConfig{
				TelemetryEnabled:      boolPtr(true),
				TelemetryOtlpEndpoint: "",
			},
			expectInit: true, // Would attempt to init
		},
		{
			name: "Disabled with endpoint",
			cfg: &config.CliConfig{
				TelemetryEnabled:      boolPtr(false),
				TelemetryOtlpEndpoint: "http://localhost:4317",
			},
			expectInit: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset tp for each test
			originalTp := tp
			tp = nil
			defer func() { tp = originalTp }()

			// This should not panic regardless of configuration
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("InitializeTelemetry panicked with config %+v: %v", tc.cfg, r)
				}
			}()

			InitializeTelemetry(tc.cfg)

			// For disabled cases, verify tp remains nil
			if !tc.expectInit && tp != nil {
				t.Errorf("Expected tp to be nil for disabled telemetry, but it was initialized")
			}
		})
	}
}

// Helper function for creating bool pointers
func boolPtr(b bool) *bool {
	return &b
}