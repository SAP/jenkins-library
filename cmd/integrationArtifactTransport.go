package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"errors"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/apim"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func integrationArtifactTransport(config integrationArtifactTransportOptions, telemetryData *telemetry.CustomData) {
	httpClient := &piperhttp.Client{}
	err := runIntegrationArtifactTransport(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runIntegrationArtifactTransport(config *integrationArtifactTransportOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {
	apimData := apim.Bundle{APIServiceKey: config.CasServiceKey, Client: httpClient}
	err := apim.Utils.InitAPIM(&apimData)
	if err != nil {
		return err
	}
	return CreateIntegrationArtifactTransportRequest(config, apimData)
}

// CreateIntegrationArtifactTransportRequest - Create a transport request for Integration Package
func CreateIntegrationArtifactTransportRequest(config *integrationArtifactTransportOptions, apistruct apim.Bundle) error {
	httpMethod := http.MethodPost
	httpClient := apistruct.Client
	createTransportRequestURL := fmt.Sprintf("%s/v1/contentResources/export", apistruct.Host)
	header := make(http.Header)
	header.Add("content-type", "application/json")
	payload, jsonError := GetCPITransportReqPayload(config)
	if jsonError != nil {
		return fmt.Errorf("Failed to get json payload for file %v, failed with error: %w", config.IntegrationPackageID, jsonError)
	}

	createTransportRequestResp, httpErr := httpClient.SendRequest(httpMethod, createTransportRequestURL, payload, header, nil)

	if httpErr != nil {
		return fmt.Errorf("HTTP %v request to %v failed with error: %w", httpMethod, createTransportRequestURL, httpErr)
	}

	if createTransportRequestResp != nil && createTransportRequestResp.Body != nil {
		defer createTransportRequestResp.Body.Close()
	}

	if createTransportRequestResp == nil {
		return fmt.Errorf("did not retrieve a HTTP response")
	}

	if createTransportRequestResp.StatusCode == http.StatusAccepted {
		log.Entry().
			WithField("IntegrationPackageID", config.IntegrationPackageID).
			Info("successfully created the integration package transport request")

		bodyText, readErr := io.ReadAll(createTransportRequestResp.Body)
		if readErr != nil {
			return fmt.Errorf("HTTP response body could not be read: %w", readErr)
		}
		jsonResponse, parsingErr := gabs.ParseJSON([]byte(bodyText))
		if parsingErr != nil {
			return fmt.Errorf("HTTP response body could not be parsed as JSON: %v: %w", string(bodyText), parsingErr)
		}
		processId := jsonResponse.Path("processId").Data().(string)

		if processId != "" {
			error := pollTransportStatus(processId, retryCount, config, httpClient, apistruct.Host)
			return error
		}
		return errors.New("Invalid process id")
	}
	responseBody, readErr := io.ReadAll(createTransportRequestResp.Body)

	if readErr != nil {
		return fmt.Errorf("HTTP response body could not be read, response status code: %v: %w", createTransportRequestResp.StatusCode, readErr)
	}
	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code : %v", string(responseBody), createTransportRequestResp.StatusCode)
	return fmt.Errorf("integration flow deployment failed, response Status code: %v", createTransportRequestResp.StatusCode)
}

// pollTransportStatus - Poll the integration package transport processing, return status or error details
func pollTransportStatus(processId string, remainingRetries int, config *integrationArtifactTransportOptions, httpClient piperhttp.Sender, apiHost string) error {

	if remainingRetries <= 0 {
		return errors.New("failed to start integration artifact after retrying several times")
	}
	transportStatus, err := getIntegrationTransportProcessingStatus(config, httpClient, apiHost, processId)
	if err != nil {
		return err
	}

	//with specific delay between each retry
	if (transportStatus == "RUNNING") || (transportStatus == "INITIAL") {
		// Calling Sleep method
		sleepTime := int(retryCount * 3)
		time.Sleep(time.Duration(sleepTime) * time.Second)
		remainingRetries--
		return pollTransportStatus(processId, retryCount, config, httpClient, apiHost)
	}

	//if artifact transport completed, then just return
	if transportStatus == "FINISHED" {
		return nil
	}

	//if error return immediately with error details
	if transportStatus == "ERROR" || transportStatus == "ABORTED" {
		resp, err := getIntegrationTransportError(config, httpClient, apiHost, processId)
		if err != nil {
			return err
		}
		return errors.New(resp)
	}
	return nil
}

// GetJSONPayload -return http payload as byte array
func GetCPITransportReqPayload(config *integrationArtifactTransportOptions) (*bytes.Buffer, error) {
	jsonObj := gabs.New()
	jsonObj.Set(rand.Intn(5000), "id")
	jsonObj.Set("MonitoringTeam", "requestor")
	jsonObj.Set("1.0.0", "version")
	jsonObj.Set("TransportManagementService", "exportMode")
	jsonObj.Set("MTAR", "exportMediaType")
	jsonObj.Set("Integration Artifact transport request for TransportManagementService", "description")
	jsonResourceObj := gabs.New()
	jsonResourceObj.Set(config.IntegrationPackageID, "id")
	jsonResourceObj.Set(config.ResourceID, "resourceID")
	jsonResourceObj.Set("d9c3fe08ceeb47a2991e53049f2ed766", "contentType")
	jsonResourceObj.Set("package", "subType")
	jsonResourceObj.Set(config.Name, "name")
	jsonResourceObj.Set("CloudIntegration", "type")
	jsonResourceObj.Set(config.Version, "version")
	jsonObj.ArrayAppend(jsonResourceObj, "contentResources")

	jsonBody, jsonErr := json.Marshal(jsonObj)

	if jsonErr != nil {
		return nil, fmt.Errorf("Transport request payload is invalid for integration package artifact %q: %w", config.IntegrationPackageID, jsonErr)
	}
	return bytes.NewBuffer(jsonBody), nil
}

// getIntegrationTransportProcessingStatus - Get integration package transport request processing Status
func getIntegrationTransportProcessingStatus(config *integrationArtifactTransportOptions, httpClient piperhttp.Sender, apiHost string, processId string) (string, error) {
	httpMethod := "GET"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	header.Add("Accept", "application/json")
	transportProcStatusURL := fmt.Sprintf("%s/v1/operations/%s", apiHost, processId)
	transportProcStatusResp, httpErr := httpClient.SendRequest(httpMethod, transportProcStatusURL, nil, header, nil)

	if transportProcStatusResp != nil && transportProcStatusResp.Body != nil {
		defer transportProcStatusResp.Body.Close()
	}

	if transportProcStatusResp == nil {
		return "", fmt.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if (transportProcStatusResp.StatusCode == http.StatusOK) || (transportProcStatusResp.StatusCode == http.StatusAccepted) {
		log.Entry().
			WithField("IntegrationPackageID", config.IntegrationPackageID).
			Info("successfully processed the integration package transport response status")

		bodyText, readErr := io.ReadAll(transportProcStatusResp.Body)
		if readErr != nil {
			return "", fmt.Errorf("HTTP response body could not be read, response status code: %v: %w", transportProcStatusResp.StatusCode, readErr)
		}
		jsonResponse, parsingErr := gabs.ParseJSON([]byte(bodyText))
		if parsingErr != nil {
			return "", fmt.Errorf("HTTP response body could not be parsed as JSON: %v: %w", string(bodyText), parsingErr)
		}
		contentTransporStatus := jsonResponse.Path("state").Data().(string)
		return contentTransporStatus, nil
	}
	if httpErr != nil {
		return getHTTPErrorMessage(httpErr, transportProcStatusResp, httpMethod, transportProcStatusURL)
	}
	return "", fmt.Errorf("failed to get transport request processing status, response Status code: %v", transportProcStatusResp.StatusCode)
}

// getTransportError - Get integration package transport failures error details
func getIntegrationTransportError(config *integrationArtifactTransportOptions, httpClient piperhttp.Sender, apiHost string, processId string) (string, error) {
	httpMethod := "GET"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	errorStatusURL := fmt.Sprintf("%s/v1/operations/%s/logs", apiHost, processId)
	errorStatusResp, httpErr := httpClient.SendRequest(httpMethod, errorStatusURL, nil, header, nil)

	if errorStatusResp != nil && errorStatusResp.Body != nil {
		defer errorStatusResp.Body.Close()
	}

	if errorStatusResp == nil {
		return "", fmt.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if errorStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("IntegrationPackageId", config.IntegrationPackageID).
			Info("Successfully retrieved deployment failures error details")
		responseBody, readErr := io.ReadAll(errorStatusResp.Body)
		if readErr != nil {
			return "", fmt.Errorf("HTTP response body could not be read, response status code: %v: %w", errorStatusResp.StatusCode, readErr)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(responseBody), errorStatusResp.StatusCode)
		errorDetails := string(responseBody)
		return errorDetails, nil
	}
	if httpErr != nil {
		return getHTTPErrorMessage(httpErr, errorStatusResp, httpMethod, errorStatusURL)
	}
	return "", fmt.Errorf("failed to get Integration Package transport error details, response Status code: %v", errorStatusResp.StatusCode)
}
