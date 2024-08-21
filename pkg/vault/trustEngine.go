package vault

import (
	"encoding/json"
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"net/http"
	"net/url"
	"time"
)

type trustEngineUtils interface {
	*piperhttp.Client
}

type TrustEngineSecret struct {
	Token                          string                         `json:"token"`
	RequestId                      string                         `json:"request_id"`
	PipelineRepositoryRelationship PipelineRepositoryRelationship `json:"pipeline_repository_relationship"`
}

type PipelineRepositoryRelationship struct {
	Name            string `json:"name"`
	PipelineGroupId string `json:"pipeline_group_id"`
	Repo            struct {
		Id  int    `json:"id"`
		Url string `json:"url"`
	} `json:"repo"`
	Sonar struct {
		Host       string `json:"host"`
		ProjectKey string `json:"project_key"`
	} `json:"sonar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetTrustEngineSecret(baseURL *url.URL, refName, jwt string, client *piperhttp.Client) (TrustEngineSecret, error) {
	var trust TrustEngineSecret
	fullURL, _ := url.JoinPath(baseURL.String(), refName)

	var header http.Header = map[string][]string{"Authorization": {fmt.Sprintf("Bearer %s", jwt)}}
	response, err := client.SendRequest("GET", fullURL, nil, header, nil)
	if err != nil {
		return trust, err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&trust)
	if err != nil {
		return trust, err
	}
	return trust, nil
}
