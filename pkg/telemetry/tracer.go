package telemetry

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func InitTracer(resAttributes []attribute.KeyValue, enabled bool) (*trace.TracerProvider, error) {
	if !enabled {
		log.Entry().Debug("tracability disabled")
		return nil, nil
	}
	log.Entry().Info("tracability enabled")

	var tracerProvider *trace.TracerProvider
	var err error

	resAttributes = append(resAttributes, semconv.ServiceName("piper-go"))
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		resAttributes...,
	//TODO use detectors
	)

	if url, ok := os.LookupEnv("OTEL_EXPORTER_JAEGER_ENDPOINT"); ok {
		tracerProvider, err = initJaegerTracer(res, url)
	} else if _, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT"); ok {
		//TODO handle OTLP
	} else {
		// tracerProvider, err = initFileTracer(res)
		// tracerProvider, _ = initStdoutTracer(res)
		tracerProvider = trace.NewTracerProvider()
	}
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tracerProvider)
	return tracerProvider, nil
}

// func initStdoutTracer(res *resource.Resource) (*trace.TracerProvider, error) {
// 	exporter, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())

// 	return trace.NewTracerProvider(
// 		trace.WithBatcher(exporter),
// 		trace.WithResource(res),
// 	), nil
// }

// func initFileTracer(res *resource.Resource) (*trace.TracerProvider, error) {
// 	f, _ := os.Create("traces.json")

// 	exporter, _ := stdouttrace.New(
// 		stdouttrace.WithWriter(f),
// 		stdouttrace.WithPrettyPrint(),
// 	)

// 	return trace.NewTracerProvider(
// 		trace.WithBatcher(exporter),
// 		trace.WithSampler(trace.AlwaysSample()),
// 		trace.WithResource(res),
// 	), nil
// }

func initJaegerTracer(res *resource.Resource, url string) (*trace.TracerProvider, error) {
	log.Entry().Infof("initializing tracing to Jaeger (%s)", url)
	// Create the Jaeger exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		log.Entry().WithError(err).Error("failed to set up tracing")
		return nil, err
	}

	return trace.NewTracerProvider(
		// Always be sure to batch in production.
		trace.WithBatcher(exporter),
		// Record information about this application in a Resource.
		trace.WithResource(res),
	), nil
}
