package telemetry

import (
	"context"
	"encoding/json"
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const service_name = "piper-cli"

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

func InitOpenTelemetry(ctx context.Context) (*sdktrace.TracerProvider, context.Context, func()) {
	ctx = restoreParent(ctx)
	for _, init := range initFunctions {
		if ok := init(); ok {
			break
		}
	}

	tp, cleanup, err := InitTracer(ctx)
	if err != nil {
		log.Entry().Info("failed to initialize OpenTelemetry")
	}

	return tp, context.WithValue(ctx, tracerKey, otel.Tracer("com.sap.piper")), cleanup
}

func restoreParent(ctx context.Context) context.Context {
	if carrierJSONString, ok := os.LookupEnv("PIPER_otel_carrier"); ok {
		var carrier propagation.MapCarrier
		if err := json.Unmarshal([]byte(carrierJSONString), &carrier); err != nil {
			log.Entry().Errorf("Failed to unmarshal carrier JSON: %v", err)
			return ctx
		}
		log.Entry().Infof("Detected parent trace %s", carrierJSONString)
		return propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		).Extract(ctx, carrier)
	}
	return ctx
}

func GetTracer(ctx context.Context) trace.Tracer {
	//TODO: handle missing tracer
	return ctx.Value(tracerKey).(trace.Tracer)
}
