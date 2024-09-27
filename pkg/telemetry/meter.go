package telemetry

// import (
// 	"context"
// 	"os"

// 	"github.com/SAP/jenkins-library/pkg/log"
// 	"go.opentelemetry.io/otel"
// 	"go.opentelemetry.io/otel/attribute"

// 	// "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"

// 	// "go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
// 	"go.opentelemetry.io/otel/metric"
// 	"go.opentelemetry.io/otel/sdk/metric/controller/basic"
// 	"go.opentelemetry.io/otel/sdk/resource"
// 	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
// )

// func InitMeter(ctx context.Context, resAttributes []attribute.KeyValue) (func(context.Context) error, error) {
// 	var err error
// 	var meterProvider *metric.MeterProvider
// 	resAttributes = append(resAttributes, semconv.ServiceName("piper-go"))
// 	res := resource.NewWithAttributes(
// 		semconv.SchemaURL,
// 		resAttributes...,
// 	)

// 	if dsn, ok := os.LookupEnv("UPTRACE_DSN"); ok {
// 		prepareUptraceMeter(ctx, res, dsn)
// 	} else if token, ok := os.LookupEnv("LIGHTSTEP_TOKEN"); ok {
// 		prepareLightstepMeter(ctx, res, token)
// 	} else if token, ok := os.LookupEnv("TELEMETRYHUB_TOKEN"); ok {
// 		prepareTelemetryHubMeter(ctx, res, token)
// 	}

// 	if _, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"); ok {
// 		meterProvider, err = initGRPCMeter(ctx, res)
// 		// } else {
// 		// 	meterProvider, err = initStdoutMeter(ctx, res)
// 	}

// 	if err != nil {
// 		return nil, err
// 	}
// 	otel.SetMeterProvider(*meterProvider)
// 	return *meterProvider.Shutdown, nil
// }

// // Inits metric reporting to https://app.uptrace.dev/
// // func initUptraceMeter(_ context.Context, res *resource.Resource) (func(context.Context) error, error) {
// // 	log.Entry().Debug("initializing metering to Uptrace")
// // 	//FIXME: runs with context.TODO(), use ctx from cmd
// // 	uptrace.ConfigureOpentelemetry(
// // 		uptrace.WithTracingDisabled(), // only init otel for metrics
// // 		uptrace.WithMetricsEnabled(true),
// // 		uptrace.WithResource(res),
// // 	)
// // 	return uptrace.Shutdown, nil
// // }

// // Inits metric reporting to https://app.uptrace.dev/
// func prepareUptraceMeter(ctx context.Context, res *resource.Resource, dsn string) {
// 	// 	otlpmetricgrpc.WithCompressor(gzip.Name),
// 	// 	otlpmetricgrpc.WithTemporalitySelector(preferDeltaTemporalitySelector),
// 	log.Entry().Debug("preparing metering to Uptrace")
// 	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "https://otlp.uptrace.dev:4317")
// 	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_HEADERS", "uptrace-dsn="+dsn)
// }

// // Inits metric reporting to https://app.lightstep.com/
// func prepareLightstepMeter(ctx context.Context, res *resource.Resource, token string) {
// 	log.Entry().Debug("preparing metering to Lightstep")
// 	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "https://ingest.lightstep.com:443")
// 	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_HEADERS", "lightstep-access-token="+token)
// }

// // Inits metric reporting to https://app.telemetryhub.com/
// func prepareTelemetryHubMeter(ctx context.Context, res *resource.Resource, token string) {
// 	log.Entry().Debug("preparing metering to TelemetryHub")
// 	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "https://otlp.telemetryhub.com:4317")
// 	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_HEADERS", "x-telemetryhub-key="+token)
// }

// func initGRPCMeter(ctx context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
// 	log.Entry().Debugf("initializing metering to %s", os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"))
// 	// 	u, _ := url.Parse(endpoint)
// 	// 	if u.Scheme == "https" {
// 	// 		// Create credentials using system certificates.
// 	// 		creds := credentials.NewClientTLSFromCert(nil, "")
// 	// 		options = append(options, otlpmetricgrpc.WithTLSCredentials(creds))
// 	// 	} else {
// 	// 		options = append(options, otlpmetricgrpc.WithInsecure())
// 	// 	}

// 	// options := []otlpmetricgrpc.Option{
// 	// 	// otlpmetricgrpc.WithInsecure(),
// 	// 	otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
// 	// }

// 	// exporter, err := otlpmetricgrpc.New(ctx, options...)
// 	// if err != nil {
// 	// 	log.Entry().WithError(err).Error("failed to initialize exporter")
// 	// 	return nil, errors.Wrap(err, "failed to initialize exporter")
// 	// }

// 	direct.New()

// 	batcher := direct.New()
// 	pusher := basic.New(batcher, basic.WithResource(res))

// 	return pusher.MeterProvider(), nil

// 	// return metric.NewMeterProvider(
// 	// 	// use large interval to only report once on shutdown
// 	// 	metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(time.Hour*24))),
// 	// 	metric.WithResource(res),
// 	// ), nil
// }

// // func initStdoutMeter(_ context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
// // 	log.Entry().Debug("initializing metering to stdout")
// // 	encoder := json.NewEncoder(os.Stdout)
// // 	encoder.SetIndent("", "  ")
// // 	exporter, err := stdoutmetric.New(stdoutmetric.WithEncoder(encoder))
// // 	if err != nil {
// // 		log.Entry().WithError(err).Warning("failed to initialize exporter")
// // 		return nil, errors.Wrap(err, "failed to initialize exporter")
// // 	}

// // 	return metric.NewMeterProvider(
// // 		metric.WithReader(metric.NewPeriodicReader(exporter)),
// // 		metric.WithResource(res),
// // 	), nil
// // }
