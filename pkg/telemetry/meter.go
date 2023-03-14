package telemetry

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc/credentials"
)

func InitMeter(ctx context.Context, resAttributes []attribute.KeyValue) (func(context.Context) error, error) {
	var err error
	var meterProvider *metric.MeterProvider
	resAttributes = append(resAttributes, semconv.ServiceName("piper-go"))
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		resAttributes...,
	)

	if _, ok := os.LookupEnv("UPTRACE_DSN"); ok {
		return initUptraceMeter(ctx, res)
	} else if token, ok := os.LookupEnv("LIGHTSTEP_TOKEN"); ok {
		meterProvider, err = initLightstepMeter(ctx, res, token)
	} else if token, ok := os.LookupEnv("TELEMETRYHUB_TOKEN"); ok {
		meterProvider, err = initTelemetryHubMeter(ctx, res, token)
	} else {
		meterProvider, err = initStdoutMeter(ctx, res)
	}
	if err != nil {
		return nil, err
	}
	global.SetMeterProvider(meterProvider)
	return meterProvider.Shutdown, nil
}

// Inits metric reporting to https://app.uptrace.dev/
func initUptraceMeter(_ context.Context, res *resource.Resource) (func(context.Context) error, error) {
	log.Entry().Debug("initializing metering to Uptrace")
	//FIXME: runs with context.TODO(), use ctx from cmd
	uptrace.ConfigureOpentelemetry(
		uptrace.WithTracingDisabled(), // only init otel for metrics
		uptrace.WithMetricsEnabled(true),
		uptrace.WithResource(res),
	)
	return uptrace.Shutdown, nil
}

// Inits metric reporting to https://app.lightstep.com/
func initLightstepMeter(ctx context.Context, res *resource.Resource, token string) (*metric.MeterProvider, error) {
	log.Entry().Debug("initializing metering to Lightstep")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "https://ingest.lightstep.com:443")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_HEADERS", "lightstep-access-token="+token)
	return initGRPCMeter(ctx, res)
}

// Inits metric reporting to https://app.telemetryhub.com/
func initTelemetryHubMeter(ctx context.Context, res *resource.Resource, token string) (*metric.MeterProvider, error) {
	log.Entry().Debug("initializing metering to TelemetryHub")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "https://otlp.telemetryhub.com:4317")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_HEADERS", "x-telemetryhub-key="+token)
	return initGRPCMeter(ctx, res)
}

func initGRPCMeter(ctx context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
	// 	u, _ := url.Parse(endpoint)
	// 	if u.Scheme == "https" {
	// 		// Create credentials using system certificates.
	// 		creds := credentials.NewClientTLSFromCert(nil, "")
	// 		options = append(options, otlpmetricgrpc.WithTLSCredentials(creds))
	// 	} else {
	// 		options = append(options, otlpmetricgrpc.WithInsecure())
	// 	}

	options := []otlpmetricgrpc.Option{
		// otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
	}

	exporter, err := otlpmetricgrpc.New(ctx, options...)
	if err != nil {
		log.Entry().WithError(err).Error("failed to initialize exporter")
		return nil, errors.Wrap(err, "failed to initialize exporter")
	}

	return metric.NewMeterProvider(
		// use large interval to only report once on shutdown
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(time.Hour*24))),
		metric.WithResource(res),
	), nil
}

func initStdoutMeter(_ context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
	log.Entry().Debug("initializing metering to stdout")
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	exporter, err := stdoutmetric.New(stdoutmetric.WithEncoder(encoder))
	if err != nil {
		log.Entry().WithError(err).Warning("failed to initialize exporter")
		return nil, errors.Wrap(err, "failed to initialize exporter")
	}

	return metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	), nil
}
