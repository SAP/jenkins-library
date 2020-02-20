package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type artifactDescription struct {
	ID         string `json:"id"`
	Classifier string `json:"classifier"`
	Type       string `json:"type"`
	File       string `json:"file"`
}

func nexusUpload(config nexusUploadOptions, telemetryData *telemetry.CustomData) error {
	log.Entry().Info(config)

	log.Entry().Info("JSON string is ", config.Artifacts)

	var artifacts []artifactDescription
	err := json.Unmarshal([]byte(config.Artifacts), &artifacts)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert JSON: %s", err))
	}

	client := piperHttp.Client{}
	clientOptions := piperHttp.ClientOptions{Username: config.User, Password: config.Password, Logger: log.Entry().WithField("package", "github.com/SAP/jenkins-library/pkg/http")}
	client.SetOptions(clientOptions)

	for _, artifact := range artifacts {
		groupPath := strings.ReplaceAll(config.groupId, ".", "/")
		artifactName := artifact.ID + "-" config.Version + "." + artifact.Type
		url := "http://" + config.Url + "/repository/" + config.Repository + "/" + groupPath + "/" + artifact.ID + "/" + config.Version + "/" + artifactName
		url = strings.ReplaceAll(url, "//", "/")

		log.Entry().Info("Trying to upload ", artifact.File, " to ", url)

		response, err := client.UploadRequest(http.MethodPut, url, artifact.File, "", nil, nil)
		if err != nil {
			if response.StatusCode == 400 {
				log.Entry().Info("Artifact already exits, deleting and retrying...")
				response, err = client.SendRequest(http.MethodDelete, url, nil, nil, nil)
				if err != nil {
					panic(fmt.Sprintf("Failed to delete artifact: %s", err))
				} else {
					response, err = client.UploadRequest(http.MethodPut, url, artifact.File, "", nil, nil)
				}
			}
			if err != nil {
				//				log.Entry().Info("Failed to upload artifact")
				panic(fmt.Sprintf("Failed to upload artifact: %s", err))
			}
		}

		log.Entry().Info("Response is ", response)
	}

	//log.Entry().WithField("customKey", "customValue").Info("This is how you write a log message with a custom field ...")
	return err
}
