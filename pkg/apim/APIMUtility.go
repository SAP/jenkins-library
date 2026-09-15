package apim

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/xsuaa"

	"github.com/google/go-querystring/query"
)

// Utils for apim
type Utils interface {
	InitAPIM() error
	IsPayloadJSON() bool
}

// OdataUtils for apim
type OdataUtils interface {
	MakeOdataQuery() (string, error)
}

// OdataParameters struct
type OdataParameters struct {
	Filter  string `url:"filter,omitempty"`
	Search  string `url:"search,omitempty"`
	Top     int    `url:"top,omitempty"`
	Skip    int    `url:"skip,omitempty"`
	Orderby string `url:"orderby,omitempty"`
	Select  string `url:"select,omitempty"`
	Expand  string `url:"expand,omitempty"`
}

// Bundle struct
type Bundle struct {
	APIServiceKey, Host, Payload string
	Client                       piperhttp.Sender
}

// InitAPIM() fumnction initialize APIM bearer token for API access
func (apim *Bundle) InitAPIM() error {
	serviceKey, err := cpi.ReadCpiServiceKey(apim.APIServiceKey)
	if err != nil {
		return err
	}
	apim.Host = serviceKey.OAuth.Host
	httpClient := apim.Client
	clientOptions := piperhttp.ClientOptions{}
	x := xsuaa.XSUAA{
		OAuthURL:     serviceKey.OAuth.OAuthTokenProviderURL,
		ClientID:     serviceKey.OAuth.ClientID,
		ClientSecret: serviceKey.OAuth.ClientSecret,
	}
	token, tokenErr := x.GetBearerToken()

	if tokenErr != nil {
		return fmt.Errorf("failed to fetch Bearer Token: %w", tokenErr)
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token.AccessToken)
	httpClient.SetOptions(clientOptions)
	return nil
}

// IsJSON checks given string is valid json or not
func (apim *Bundle) IsPayloadJSON() bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(apim.Payload), &js) == nil
}

func (odataFilters *OdataParameters) MakeOdataQuery() (string, error) {

	values, encodeErr := query.Values(odataFilters)
	if encodeErr != nil {
		return "", encodeErr
	}
	encoded := values.Encode()
	if len(encoded) > 0 {
		encoded = "?" + strings.ReplaceAll(encoded, "&", "&$")
	}
	return encoded, nil
}
