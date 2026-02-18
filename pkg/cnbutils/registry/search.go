// Package registry provides utilities to search buildpacks using registry API
package registry

import (
	"encoding/json"
	"fmt"
	"net/http"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	defaultRegistryAPI = "https://registry.buildpacks.io/api/v1/buildpacks"
)

type latest struct {
	Version string         `json:"version"`
	Misc    map[string]any `json:"-"`
}

type version struct {
	Version string `json:"version"`
	Link    string `json:"_link"`
}

type response struct {
	Latest   latest    `json:"latest"`
	Versions []version `json:"versions"`
}

type versionResponse struct {
	Addr string         `json:"addr"`
	Misc map[string]any `json:"-"`
}

func SearchBuildpack(id, version string, httpClient piperhttp.Sender, baseApiURL string) (string, error) {
	var apiResponse response

	if baseApiURL == "" {
		baseApiURL = defaultRegistryAPI
	}

	apiURL := fmt.Sprintf("%s/%s", baseApiURL, id)

	rawResponse, err := httpClient.SendRequest(http.MethodGet, apiURL, nil, nil, nil)
	if err != nil {
		return "", err
	}
	defer rawResponse.Body.Close()

	err = json.NewDecoder(rawResponse.Body).Decode(&apiResponse)
	if err != nil {
		return "", fmt.Errorf("unable to parse response from the %s, error: %s", apiURL, err.Error())
	}

	if version == "" {
		version = apiResponse.Latest.Version
		log.Entry().Infof("Version for the buildpack '%s' is not specified, using the latest '%s'", id, version)
	}

	for _, ver := range apiResponse.Versions {
		if ver.Version == version {
			return getImageAddr(ver.Link, httpClient)
		}
	}

	return "", fmt.Errorf("version '%s' was not found for the buildpack '%s'", version, id)
}

func getImageAddr(link string, httpClient piperhttp.Sender) (string, error) {
	var verResponse versionResponse

	rawResponse, err := httpClient.SendRequest(http.MethodGet, link, nil, nil, nil)
	if err != nil {
		return "", err
	}
	defer rawResponse.Body.Close()

	err = json.NewDecoder(rawResponse.Body).Decode(&verResponse)
	if err != nil {
		return "", fmt.Errorf("unable to parse response from the %s, error: %s", link, err.Error())
	}

	return verResponse.Addr, nil
}
