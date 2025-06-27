package telemetry

import (
	"context"
	"fmt"
	"gemini-cli-go/internal/config"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var tp *sdktrace.TracerProvider

// InitializeTelemetry initializes the OpenTelemetry SDK.
func InitializeTelemetry(cfg *config.CliConfig) {
	if cfg.TelemetryEnabled == nil || !*cfg.TelemetryEnabled {
		fmt.Println("Telemetry is disabled.")
		return
	}

	fmt.Println("Initializing Telemetry...")

	ctx := context.Background()

	// Create a new OTLP gRPC exporter
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(cfg.TelemetryOtlpEndpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to create OTLP exporter: %v", err)
	}

	// Create a new tracer provider with the exporter
	tp = sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("gemini-cli-go"),
		)),
	)

	// Set the global TracerProvider
	otel.SetTracerProvider(tp)

	fmt.Println("Telemetry initialized.")
}

// ShutdownTelemetry shuts down the OpenTelemetry SDK.
func ShutdownTelemetry(ctx context.Context) {
	if tp == nil {
		fmt.Println("Telemetry not initialized, skipping shutdown.")
		return
	}

	fmt.Println("Shutting down Telemetry...")

	if err := tp.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown TracerProvider: %v", err)
	}

	fmt.Println("Telemetry shut down.")
}
