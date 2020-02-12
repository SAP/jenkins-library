package cmd

import (
	"github.com/SAP/jenkins-library/pkg/http"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func nexusUpload(config nexusUploadOptions, telemetryData *telemetry.CustomData) error {
	log.Entry().Info(config)
	client := http.Client{}
	client.UploadFile(config.Url, "", "", nil, nil)

	log.Entry().WithField("customKey", "customValue").Info("This is how you write a log message with a custom field ...")
	return nil
}
