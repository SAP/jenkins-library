package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func getPackageList(config getPackageListOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *getPackageListCommonPipelineEnvironment) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runGetPackageList(&config, telemetryData, httpClient, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runGetPackageList(config *getPackageListOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender, commonPipelineEnvironment *getPackageListCommonPipelineEnvironment) error {
	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/json")
	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}
	servieEndpointURL := fmt.Sprintf("%s/api/v1/IntegrationPackages", serviceKey.OAuth.Host)
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := "GET"
	integrationPackageResp, httpErr := httpClient.SendRequest(httpMethod, servieEndpointURL, nil, header, nil)

	if httpErr != nil {
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error", httpMethod, servieEndpointURL)
	}

	if integrationPackageResp != nil && integrationPackageResp.Body != nil {
		defer integrationPackageResp.Body.Close()
	}

	if integrationPackageResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if integrationPackageResp.StatusCode == 200 {
		bodyText, readErr := ioutil.ReadAll(integrationPackageResp.Body)
		if readErr != nil {
			return errors.Wrap(readErr, "HTTP response body could not be read")
		}
		jsonResponse, parsingErr := gabs.ParseJSON([]byte(bodyText))
		if parsingErr != nil {
			return errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
		}
		commonPipelineEnvironment.custom.integrationPackageList += "{\"Packages\" : {"
		for _, child := range jsonResponse.S("d", "results").Children() {
			// iflowID := strings.ReplaceAll(child.Path("Name").String(), "\"", "")
			// if iflowID == config.IntegrationFlowID {
			entryPoints := child.S("Id")
			finalEndpoint := entryPoints.Data().(string)
			lastChar := commonPipelineEnvironment.custom.integrationPackageList[len(commonPipelineEnvironment.custom.integrationPackageList)]
			if lastChar == '{' {
				commonPipelineEnvironment.custom.integrationPackageList += "\n\"" + finalEndpoint + "\": {\n"
			} else {
				commonPipelineEnvironment.custom.integrationPackageList += ",\n\"" + finalEndpoint + "\": {\n"
			}
			iFlowURL := fmt.Sprintf("%s/api/v1/IntegrationPackages('%s')/IntegrationDesigntimeArtifacts", serviceKey.OAuth.Host, finalEndpoint)
			vMapURL := fmt.Sprintf("%s/api/v1/IntegrationPackages('%s')/ValueMappingDesigntimeArtifacts", serviceKey.OAuth.Host, finalEndpoint)
			mMapURL := fmt.Sprintf("%s/api/v1/IntegrationPackages('%s')/MessageMappingDesigntimeArtifacts", serviceKey.OAuth.Host, finalEndpoint)
			sCollURL := fmt.Sprintf("%s/api/v1/IntegrationPackages('%s')/ScriptCollectionDesigntimeArtifacts", serviceKey.OAuth.Host, finalEndpoint)
			iFlowResp, httpErr1 := httpClient.SendRequest(httpMethod, iFlowURL, nil, header, nil)
			vMapResp, httpErr2 := httpClient.SendRequest(httpMethod, vMapURL, nil, header, nil)
			mMapResp, httpErr3 := httpClient.SendRequest(httpMethod, mMapURL, nil, header, nil)
			sCollResp, httpErr4 := httpClient.SendRequest(httpMethod, sCollURL, nil, header, nil)
			if httpErr1 != nil && httpErr2 != nil && httpErr3 != nil && httpErr4 != nil {
				return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error", httpMethod, servieEndpointURL)
			}

			commonPipelineEnvironment.custom.integrationPackageList += "\"IntegrationDesigntimeArtifacts\": ["
			if iFlowResp.StatusCode == 200 {
				bodyText1, readErr1 := ioutil.ReadAll(iFlowResp.Body)
				jsonResponse1, parsingErr1 := gabs.ParseJSON([]byte(bodyText1))
				if readErr1 != nil {
					return errors.Wrap(readErr1, "HTTP response body could not be read")
				}
				if parsingErr1 != nil {
					return errors.Wrapf(parsingErr1, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
				}
				for _, child1 := range jsonResponse1.S("d", "results").Children() {
					entryPoints1 := child1.S("Id")
					finalEndpoint1 := entryPoints1.Data().(string)
					lastChar2 := commonPipelineEnvironment.custom.integrationPackageList[len(commonPipelineEnvironment.custom.integrationPackageList)]
					if lastChar2 == '[' {
						commonPipelineEnvironment.custom.integrationPackageList += "\n\"" + finalEndpoint1 + "\""
					} else {
						commonPipelineEnvironment.custom.integrationPackageList += ",\n\"" + finalEndpoint1 + "\""
					}
				}
			}
			commonPipelineEnvironment.custom.integrationPackageList += "],\n\"ValueMappingDesigntimeArtifacts\": [\n"
			if vMapResp.StatusCode == 200 {
				bodyText2, readErr2 := ioutil.ReadAll(vMapResp.Body)
				jsonResponse2, parsingErr2 := gabs.ParseJSON([]byte(bodyText2))
				if readErr2 != nil {
					return errors.Wrap(readErr2, "HTTP response body could not be read")
				}
				if parsingErr2 != nil {
					return errors.Wrapf(parsingErr2, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
				}
				for _, child2 := range jsonResponse2.S("d", "results").Children() {
					entryPoints2 := child2.S("Id")
					finalEndpoint2 := entryPoints2.Data().(string)
					lastChar3 := commonPipelineEnvironment.custom.integrationPackageList[len(commonPipelineEnvironment.custom.integrationPackageList)]
					if lastChar3 == '[' {
						commonPipelineEnvironment.custom.integrationPackageList += "\n\"" + finalEndpoint2 + "\""
					} else {
						commonPipelineEnvironment.custom.integrationPackageList += ",\n\"" + finalEndpoint2 + "\""
					}
				}
			}
			commonPipelineEnvironment.custom.integrationPackageList += "],\n\"MessageMappingDesigntimeArtifacts\": [\n"
			if mMapResp.StatusCode == 200 {
				bodyText3, readErr3 := ioutil.ReadAll(mMapResp.Body)
				jsonResponse3, parsingErr3 := gabs.ParseJSON([]byte(bodyText3))
				if readErr3 != nil {
					return errors.Wrap(readErr3, "HTTP response body could not be read")
				}
				if parsingErr3 != nil {
					return errors.Wrapf(parsingErr3, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
				}
				for _, child3 := range jsonResponse3.S("d", "results").Children() {
					entryPoints3 := child3.S("Id")
					finalEndpoint3 := entryPoints3.Data().(string)
					lastChar4 := commonPipelineEnvironment.custom.integrationPackageList[len(commonPipelineEnvironment.custom.integrationPackageList)]
					if lastChar4 == '[' {
						commonPipelineEnvironment.custom.integrationPackageList += "\n\"" + finalEndpoint3 + "\""
					} else {
						commonPipelineEnvironment.custom.integrationPackageList += ",\n\"" + finalEndpoint3 + "\""
					}
				}
			}
			commonPipelineEnvironment.custom.integrationPackageList += "],\n\"ScriptCollectionDesigntimeArtifacts\": [\n"
			if sCollResp.StatusCode == 200 {
				bodyText4, readErr4 := ioutil.ReadAll(sCollResp.Body)
				jsonResponse4, parsingErr4 := gabs.ParseJSON([]byte(bodyText4))
				if readErr4 != nil {
					return errors.Wrap(readErr4, "HTTP response body could not be read")
				}
				if parsingErr4 != nil {
					return errors.Wrapf(parsingErr4, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
				}
				for _, child4 := range jsonResponse4.S("d", "results").Children() {
					entryPoints4 := child4.S("Id")
					finalEndpoint4 := entryPoints4.Data().(string)
					lastChar5 := commonPipelineEnvironment.custom.integrationPackageList[len(commonPipelineEnvironment.custom.integrationPackageList)]
					if lastChar5 == '[' {
						commonPipelineEnvironment.custom.integrationPackageList += "\n\"" + finalEndpoint4 + "\""
					} else {
						commonPipelineEnvironment.custom.integrationPackageList += ",\n\"" + finalEndpoint4 + "\""
					}
				}
			}

			commonPipelineEnvironment.custom.integrationPackageList += "]\n}"
			// return nil

		}
		commonPipelineEnvironment.custom.integrationPackageList += "}}\n"
		return nil
	}

	responseBody, readErr := ioutil.ReadAll(integrationPackageResp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", integrationPackageResp.StatusCode)
	}

	log.Entry().Errorf("a HTTP error occurred!  Response body: %v, Response status code: %v", string(responseBody), integrationPackageResp.StatusCode)
	return errors.Errorf("Unable to get integration packages, Response Status code: %v", integrationPackageResp.StatusCode)
}
