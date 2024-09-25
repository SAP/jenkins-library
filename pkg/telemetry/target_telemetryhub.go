package telemetry

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
)

const telemetryhub_env_var = "TELEMETRYHUB_TOKEN"
const telemetryhub_endpoint = "https://otlp.telemetryhub.com:4317"
const telemetryhub_header = "x-telemetryhub-key="

// Inits reporting to https://app.telemetryhub.com
func initWithTelemetryhub() bool {
	if token, ok := os.LookupEnv(telemetryhub_env_var); ok {
		log.Entry().Info("using OpenTelemetry with TelemetryHub")
		if err := os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", telemetryhub_endpoint); err != nil {
			log.Entry().Infof("Error setting env var: %s", err)
		}
		if err := os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", telemetryhub_header+token); err != nil {
			log.Entry().Infof("Error setting env var: %s", err)
		}
		// OTEL_EXPORTER_OTLP_ENDPOINT
		os.Setenv("OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION", "BASE2_EXPONENTIAL_BUCKET_HISTOGRAM")
		os.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "DELTA")
		return true
	}
	return false
}
