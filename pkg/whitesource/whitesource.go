package whitesource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

// Product defines a WhiteSource product with name and token
type Product struct {
	Name           string `json:"name"`
	Token          string `json:"token"`
	CreationDate   string `json:"creationDate,omitempty"`
	LastUpdateDate string `json:"lastUpdatedDate,omitempty"`
}

// Project defines a WhiteSource project with name and token
type Project struct {
	Name           string `json:"name"`
	PluginName     string `json:"pluginName"`
	Token          string `json:"token"`
	UploadedBy     string `json:"uploadedBy"`
	CreationDate   string `json:"creationDate,omitempty"`
	LastUpdateDate string `json:"lastUpdatedDate,omitempty"`
}

// Request defines a request object to be sent to the WhiteSource system
type Request struct {
	RequestType  string `json:"requestType,omitempty"`
	UserKey      string `json:"userKey,omitempty"`
	ProductToken string `json:"productToken,omitempty"`
	ProjectToken string `json:"projectToken,omitempty"`
	OrgToken     string `json:"orgToken,omitempty"`
}

// System defines a WhiteSource system including respective tokens (e.g. org token, user token)
type System struct {
	HTTPClient piperhttp.Sender
	OrgToken   string
	ServerURL  string
	UserToken  string
}

// GetProductsMetaInfo retrieves meta information for all WhiteSource products a user has access to
func (s *System) GetProductsMetaInfo() ([]Product, error) {
	wsResponse := struct {
		ProductVitals []Product `json:"productVitals"`
	}{
		ProductVitals: []Product{},
	}

	req := Request{
		RequestType: "getOrganizationProductVitals",
	}

	respBody, err := s.sendRequest(req)
	if err != nil {
		return wsResponse.ProductVitals, errors.Wrap(err, "WhiteSource request failed")
	}

	err = json.Unmarshal(respBody, &wsResponse)
	if err != nil {
		return wsResponse.ProductVitals, errors.Wrap(err, "failed to parse WhiteSource response")
	}

	return wsResponse.ProductVitals, nil
}

// GetMetaInfoForProduct retrieves meta information for a specific WhiteSource product
func (s *System) GetMetaInfoForProduct(productName string) (Product, error) {
	products, err := s.GetProductsMetaInfo()
	if err != nil {
		return Product{}, errors.Wrap(err, "failed to retrieve WhiteSource products")
	}

	for _, p := range products {
		if p.Name == productName {
			return p, nil
		}
	}

	return Product{}, fmt.Errorf("product '%v' not found in WhiteSource", productName)
}

// GetProjectsMetaInfo retrieves meta information for a specific WhiteSource product
func (s *System) GetProjectsMetaInfo(productToken string) ([]Project, error) {
	wsResponse := struct {
		ProjectVitals []Project `json:"projectVitals"`
	}{
		ProjectVitals: []Project{},
	}

	req := Request{
		RequestType:  "getProductProjectVitals",
		ProductToken: productToken,
	}

	respBody, err := s.sendRequest(req)
	if err != nil {
		return wsResponse.ProjectVitals, errors.Wrap(err, "WhiteSource request failed")
	}

	err = json.Unmarshal(respBody, &wsResponse)
	if err != nil {
		return wsResponse.ProjectVitals, errors.Wrap(err, "failed to parse WhiteSource response")
	}

	return wsResponse.ProjectVitals, nil
}

//GetProjectToken returns the project token for a project with a given name
func (s *System) GetProjectToken(productToken, projectName string) (string, error) {
	projects, err := s.GetProjectsMetaInfo(productToken)
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve WhiteSource project meta info")
	}

	for _, project := range projects {
		if projectName == project.Name {
			return project.Token, nil
		}
	}
	return "", nil
}

//GetProjectTokens returns the project tokens for a list of given project names
func (s *System) GetProjectTokens(productToken string, projectNames []string) ([]string, error) {
	projectTokens := []string{}

	projects, err := s.GetProjectsMetaInfo(productToken)
	if err != nil {
		return projectTokens, errors.Wrap(err, "failed to retrieve WhiteSource project meta info")
	}

	for _, project := range projects {
		for _, projectName := range projectNames {
			if projectName == project.Name {
				projectTokens = append(projectTokens, project.Token)
			}
		}
	}
	return projectTokens, nil
}

//GetProductName returns the product name for a given product token
func (s *System) GetProductName(productToken string) (string, error) {
	wsResponse := struct {
		ProductTags []Product `json:"productTags"`
	}{
		ProductTags: []Product{},
	}

	req := Request{
		RequestType:  "getProductTags",
		ProductToken: productToken,
	}

	respBody, err := s.sendRequest(req)
	if err != nil {
		return "", errors.Wrap(err, "WhiteSource request failed")
	}

	err = json.Unmarshal(respBody, &wsResponse)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse WhiteSource response")
	}

	if len(wsResponse.ProductTags) > 0 {
		return wsResponse.ProductTags[0].Name, nil
	}
	return "", nil
}

func (s *System) sendRequest(req Request) ([]byte, error) {
	responseBody := []byte{}
	if req.UserKey == "" {
		req.UserKey = s.UserToken
	}
	if req.OrgToken == "" {
		req.OrgToken = s.OrgToken
	}

	body, err := json.Marshal(req)
	if err != nil {
		return responseBody, errors.Wrap(err, "failed to create WhiteSource request")
	}

	log.Entry().Debug(string(body))

	headers := http.Header{}
	headers.Add("Content-Type", "application/json")
	response, err := s.HTTPClient.SendRequest(http.MethodPost, s.ServerURL, bytes.NewBuffer(body), headers, nil)

	if err != nil {
		return responseBody, errors.Wrap(err, "failed to send request to WhiteSource")
	}

	responseBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return responseBody, errors.Wrap(err, "failed to read WhiteSource response")
	}
	return responseBody, nil
}
