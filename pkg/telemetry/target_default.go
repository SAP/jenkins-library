package telemetry

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
)

const default_env_var = "OTEL_EXPORTER_OTLP_ENDPOINT"

// Inits reporting
func initDefault() bool {
	if url, ok := os.LookupEnv(default_env_var); ok {
		log.Entry().Infof("using OpenTelemetry with %s", url)
		// OTEL_EXPORTER_OTLP_ENDPOINT
		os.Setenv("OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION", "BASE2_EXPONENTIAL_BUCKET_HISTOGRAM")
		os.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "DELTA")
		return true
	}
	return false
}
