package telemetry

import (
	"context"

	"github.com/SAP/jenkins-library/pkg/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const service_name = "Piper"

type key struct {
	id string
}

var tracerKey = key{id: "piper"}

var initFunctions = []func() bool{
	initDefault, // check if otel envvar is already set
	initWithUptrace,
	initWithLightstep,
	initWithTelemetryhub,
	// initWithSplunk,
}

const EnvVar_otel_endpoint = ""

func InitOpenTelemetry(ctx context.Context) (context.Context, func()) {
	for _, init := range initFunctions {
		if ok := init(); ok {
			break
		}
	}

	cleanup, err := InitTracer(ctx)
	if err != nil {
		log.Entry().Info("failed to initialize OpenTelemetry")
	}

	return context.WithValue(ctx, tracerKey, otel.Tracer("com.sap.piper")), cleanup
}

func GetTracer(ctx context.Context) trace.Tracer {
	//TODO: handle missing tracer
	return ctx.Value(tracerKey).(trace.Tracer)
}
