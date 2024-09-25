package telemetry

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc/credentials"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// Inits metric reporting to https://app.uptrace.dev/
func prepareUptraceTracer(ctx context.Context, res *resource.Resource, dsn string) {
	// 	otlpmetricgrpc.WithCompressor(gzip.Name),
	// 	otlpmetricgrpc.WithTemporalitySelector(preferDeltaTemporalitySelector),
	log.Entry().Info("preparing tracing to Uptrace")
	err := os.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "https://otlp.uptrace.dev:4317")
	if err != nil {
		log.Entry().Infof("Error setting env var: %s", err)
	}
	err = os.Setenv("OTEL_EXPORTER_OTLP_TRACES_HEADERS", "uptrace-dsn="+dsn)
	if err != nil {
		log.Entry().Infof("Error setting env var: %s", err)
	}
	err = os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://otlp.uptrace.dev:4317")
	if err != nil {
		log.Entry().Infof("Error setting env var: %s", err)
	}
	err = os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "uptrace-dsn="+dsn)
	if err != nil {
		log.Entry().Infof("Error setting env var: %s", err)
	}
	// OTEL_EXPORTER_OTLP_ENDPOINT
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION", "BASE2_EXPONENTIAL_BUCKET_HISTOGRAM")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "DELTA")
}

func InitTracer(ctx context.Context, resAttributes []attribute.KeyValue) (func(), error) {
	var err error
	var tracerProvider *trace.TracerProvider

	log.Entry().Info("STARTING3")
	resAttributes = append(resAttributes, semconv.ServiceName("piper-go"))
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		resAttributes...,
	)

	if dsn, ok := os.LookupEnv("UPTRACE_DSN"); ok {
		log.Entry().Infof("STARTING4 %s", dsn)
		prepareUptraceTracer(ctx, res, dsn)
		// } else if token, ok := os.LookupEnv("LIGHTSTEP_TOKEN"); ok {
		// 	prepareLightstepMeter(ctx, res, token)
		// } else if token, ok := os.LookupEnv("TELEMETRYHUB_TOKEN"); ok {
		// 	prepareTelemetryHubMeter(ctx, res, token)
	}

	if url, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT"); ok {
		log.Entry().Infof("STARTING5 %s", url)
		tracerProvider, err = initGRPCTracer(ctx, res)
		// } else {
		// 	tracerProvider, err = initStdoutMeter(ctx, res)
	}
	if err != nil {
		return nil, err
	}

	// global.SetMeterProvider(meterProvider)
	otel.SetTracerProvider(tracerProvider)
	return func() {
		log.Entry().Infof("Shutting down TracerProvider...")
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Entry().Infof("Failed to shutdown TracerProvider: %v", err)
		}
		log.Entry().Infof("TracerProvider shut down.")
	}, nil
}

func initGRPCTracer(ctx context.Context, res *resource.Resource) (*trace.TracerProvider, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	log.Entry().Infof("initializing tracing to %s", endpoint)
	// 	u, _ := url.Parse(endpoint)
	// 	if u.Scheme == "https" {
	// 		// Create credentials using system certificates.
	// 		creds := credentials.NewClientTLSFromCert(nil, "")
	// 		options = append(options, otlpmetricgrpc.WithTLSCredentials(creds))
	// 	} else {
	// 		options = append(options, otlpmetricgrpc.WithInsecure())
	// 	}

	// options := []otlptracegrpc.Option{
	// 	// otlpmetricgrpc.WithInsecure(),
	// 	otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
	// }

	// exporter, err := otlptracegrpc.New(ctx, options...)
	// if err != nil {
	// 	log.Entry().WithError(err).Error("failed to initialize exporter")
	// 	return nil, errors.Wrap(err, "failed to initialize exporter")
	// }

	// tracerProvider := trace.NewTracerProvider(
	// 	trace.WithResource(res),
	// 	// trace.WithIDGenerator(xray.NewIDGenerator()),
	// )
	// tracerProvider.RegisterSpanProcessor(trace.NewBatchSpanProcessor(
	// 	exporter,
	// 	trace.WithMaxQueueSize(10_000),
	// 	trace.WithMaxExportBatchSize(10_000),
	// ))

	// traceExporter, err := otlptracehttp.New(ctx)
	/*traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(traceExporter)),
	)
	return traceProvider, nil*/

	creds := credentials.NewTLS(nil)

	// Creating OTLP gRPC exporter with secure connection
	log.Entry().Infof("Creating OTLP gRPC exporter...")
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("otlp.uptrace.dev:4317"),
		otlptracegrpc.WithTLSCredentials(creds),
		//otlptracegrpc.WithCompression(otlptracegrpc.GzipCompression),
		otlptracegrpc.WithTimeout(10*time.Second), // Increase timeout duration
	)
	if err != nil {
		log.Entry().Errorf("Failed to create the collector exporter: %v", err)
	}
	log.Entry().Infof("OTLP gRPC exporter created.")

	log.Entry().Infof("Creating resource...")
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("Piper"),
		semconv.ServiceVersionKey.String("v0.1.0"),
		semconv.DeploymentEnvironmentKey.String("development"),
	)

	log.Entry().Infof("Setting up TracerProvider...")
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	)

	// Set this tracer provider as global
	//otel.SetTracerProvider(tp)

	// Defer the tracerProvider's shutdown
	/*defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Entry().Warnf("failed to shutdown the tracer provider: %v", err)
		}
	}()*/

	return tp, nil
}
