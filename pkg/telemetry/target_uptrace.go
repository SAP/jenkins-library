package telemetry

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
)

const uptrace_env_var = "UPTRACE_DSN"
const uptrace_endpoint = "https://otlp.uptrace.dev:4317"
const uptrace_header = "uptrace-dsn="

// Inits reporting to https://app.uptrace.dev/
func initWithUptrace() bool {
	if dsn, ok := os.LookupEnv(uptrace_env_var); ok {
		log.Entry().Info("using OpenTelemetry with Uptrace")
		if err := os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", uptrace_endpoint); err != nil {
			log.Entry().Infof("Error setting env var: %s", err)
		}
		if err := os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", uptrace_header+dsn); err != nil {
			log.Entry().Infof("Error setting env var: %s", err)
		}
		// OTEL_EXPORTER_OTLP_ENDPOINT
		os.Setenv("OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION", "BASE2_EXPONENTIAL_BUCKET_HISTOGRAM")
		os.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "DELTA")
		return true
	}
	return false
}
