package telemetry

import (
	"context"
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc/credentials"
)

func InitTracer(ctx context.Context) (*trace.TracerProvider, func(), error) {
	var err error
	var tracerProvider *trace.TracerProvider

	//TODO: handle missing endpoint -> use stdout
	if _, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT"); ok {
		if tracerProvider, err = initGRPCTracer(ctx); err != nil {
			return nil, nil, err
		}
	}

	log.Entry().Infof("Setting up TracerProvider...")
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, func() {
		log.Entry().Infof("Shutting down TracerProvider...")
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Entry().Infof("Failed to shutdown TracerProvider: %v", err)
		}
		log.Entry().Infof("TracerProvider shut down.")
	}, nil
}

func initGRPCTracer(ctx context.Context) (*trace.TracerProvider, error) {
	log.Entry().Infof("initializing tracing to %s", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	// 	u, _ := url.Parse(endpoint)
	// 	if u.Scheme == "https" {
	// 		// Create credentials using system certificates.
	// 		creds := credentials.NewClientTLSFromCert(nil, "")
	// 		options = append(options, otlpmetricgrpc.WithTLSCredentials(creds))
	// 	} else {
	// 		options = append(options, otlpmetricgrpc.WithInsecure())
	// 	}

	// Creating OTLP gRPC exporter with secure connection
	log.Entry().Infof("Creating OTLP gRPC exporter...")
	// traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	exporter, err := otlptracegrpc.New(ctx,
		//TODO: use env var
		otlptracegrpc.WithEndpoint("otlp.uptrace.dev:4317"),
		otlptracegrpc.WithTLSCredentials(credentials.NewTLS(nil)),
		//otlptracegrpc.WithInsecure(),
		//otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
	)
	if err != nil {
		log.Entry().Errorf("Failed to create the collector exporter: %v", err)
	}
	log.Entry().Infof("OTLP gRPC exporter created.")
	return newTracerProvider(exporter), nil
}

func newTracerProvider(exporter *otlptrace.Exporter) *trace.TracerProvider {
	return trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(service_name),
				// semconv.ServiceVersionKey.String("v0.1.0"),
				// semconv.DeploymentEnvironmentKey.String("development"),
				semconv.TelemetrySDKLanguageGo,
			),
		),
	)
}
