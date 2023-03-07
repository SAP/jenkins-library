package telemetry

import (
	"context"
	"encoding/json"
	"os"

	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func InitMeter(resAttributes []attribute.KeyValue) (func(context.Context) error, error) {
	var meterProvider *metric.MeterProvider
	resAttributes = append(resAttributes, semconv.ServiceName("piper-go"))
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		resAttributes...,
	)

	if url := os.Getenv("UPTRACE_DSN"); url != "" {
		return initUptraceMeter(res, url)
	} else {
		meterProvider, _ = initStdoutMeter(res)
	}
	global.SetMeterProvider(meterProvider)
	return meterProvider.Shutdown, nil
}

func initStdoutMeter(res *resource.Resource) (*metric.MeterProvider, error) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	exporter, _ := stdoutmetric.New(stdoutmetric.WithEncoder(encoder))

	return metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	), nil
}

func initUptraceMeter(res *resource.Resource, url string) (func(context.Context) error, error) {
	uptrace.ConfigureOpentelemetry(
		uptrace.WithDSN(url),
		uptrace.WithTracingDisabled(), // only init otel fror metrics
		uptrace.WithMetricsEnabled(true),
		uptrace.WithResource(res),
		// uptrace.WithServiceVersion("1.0.0"),
	)
	return uptrace.Shutdown, nil
}
