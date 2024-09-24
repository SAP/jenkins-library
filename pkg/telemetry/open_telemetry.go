package telemetry

import (
	"context"

	"github.com/SAP/jenkins-library/pkg/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type key struct {
	id string
}

var tracerKey = key{id: "piper"}

func InitOpenTelemetry(ctx context.Context) context.Context {

	log.Entry().Info("STARTING2")
	// _, _ :=
	InitTracer(ctx, []attribute.KeyValue{})

	return context.WithValue(ctx, tracerKey, otel.Tracer("com.sap.piper"))

	// t.shutdownOpenTelemetry, err = InitMeter(t.ctx, res)
	// if err != nil {
	// 	log.Entry().WithError(err).Error("failed to initialize telemetry")
	// }

	// t.shutdownOpenTelemetryTracing, err = InitTracer(t.ctx, res)
	// if err != nil {
	// 	log.Entry().WithError(err).Error("failed to initialize telemetry (tracing)")
	// }

}

func GetTracer(ctx context.Context) trace.Tracer {
	return ctx.Value(tracerKey).(trace.Tracer)
}
