package telemetry

import (
	"context"
	"encoding/json"
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func InitMeter(resAttributes []attribute.KeyValue) (func(context.Context) error, error) {
	var err error
	var meterProvider *metric.MeterProvider
	resAttributes = append(resAttributes, semconv.ServiceName("piper-go"))
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		resAttributes...,
	)

	if _, ok := os.LookupEnv("UPTRACE_DSN"); ok {
		return initUptraceMeter(res)
	} else {
		meterProvider, err = initStdoutMeter(res)
		if err != nil {
			log.Entry().WithError(err).Warning("failed to initialize stdout telemetry meter")
			return nil, err
		}
	}
	global.SetMeterProvider(meterProvider)
	return meterProvider.Shutdown, nil
}

func initStdoutMeter(res *resource.Resource) (*metric.MeterProvider, error) {
	log.Entry().Debug("initializing metering to stdout")
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	exporter, err := stdoutmetric.New(stdoutmetric.WithEncoder(encoder))
	if err != nil {
		log.Entry().WithError(err).Warning("failed to initialize stdout telemetry meter")
		return nil, err
	}

	return metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	), nil
}

func initUptraceMeter(res *resource.Resource) (func(context.Context) error, error) {
	log.Entry().Debug("initializing metering to Uptrace")
	uptrace.ConfigureOpentelemetry(
		// uptrace.WithDSN(url), // UPTRACE_DSN is checked by default
		uptrace.WithTracingDisabled(), // only init otel fror metrics
		uptrace.WithMetricsEnabled(true),
		uptrace.WithResource(res),
		// uptrace.WithServiceVersion("1.0.0"),
	)
	return uptrace.Shutdown, nil
}
