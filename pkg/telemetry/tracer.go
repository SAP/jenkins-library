package telemetry

import (
	"context"
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// Inits metric reporting to https://app.uptrace.dev/
func prepareUptraceTracer(ctx context.Context, res *resource.Resource, dsn string) {
	// 	otlpmetricgrpc.WithCompressor(gzip.Name),
	// 	otlpmetricgrpc.WithTemporalitySelector(preferDeltaTemporalitySelector),
	log.Entry().Info("preparing tracing to Uptrace")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "https://otlp.uptrace.dev:4317")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_HEADERS", "uptrace-dsn="+dsn)
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://otlp.uptrace.dev:4317")
	os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "uptrace-dsn="+dsn)
	// OTEL_EXPORTER_OTLP_ENDPOINT
}

func InitTracer(ctx context.Context, resAttributes []attribute.KeyValue) (func(context.Context) error, error) {
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
	return tracerProvider.Shutdown, nil
}

func initGRPCTracer(ctx context.Context, res *resource.Resource) (*trace.TracerProvider, error) {
	log.Entry().Infof("initializing tracing to %s", os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"))
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
	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(traceExporter)),
	)
	return traceProvider, nil
}

func get() {

}
