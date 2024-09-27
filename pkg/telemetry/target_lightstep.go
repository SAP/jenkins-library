package telemetry

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
)

const lightstep_env_var = "LIGHTSTEP_TOKEN"
const lightstep_endpoint = "https://ingest.lightstep.com:443"
const lightstep_header = "lightstep-access-token="

// Inits reporting to https://app.lightstep.com
func initWithLightstep() bool {
	if token, ok := os.LookupEnv(lightstep_env_var); ok {
		log.Entry().Info("using OpenTelemetry with Lightstep")
		if err := os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", lightstep_endpoint); err != nil {
			log.Entry().Infof("Error setting env var: %s", err)
		}
		if err := os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", lightstep_header+token); err != nil {
			log.Entry().Infof("Error setting env var: %s", err)
		}
		// OTEL_EXPORTER_OTLP_ENDPOINT
		os.Setenv("OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION", "BASE2_EXPONENTIAL_BUCKET_HISTOGRAM")
		os.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "DELTA")
		return true
	}
	return false
}
