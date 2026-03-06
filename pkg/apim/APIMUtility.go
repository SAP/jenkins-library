package apim

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/pasztorpisti/qs"
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
	Filter, Search          string
	Top, Skip               int
	Orderby, Select, Expand string
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

	customMarshaler := qs.NewMarshaler(&qs.MarshalOptions{
		DefaultMarshalPresence: qs.OmitEmpty,
	})
	values, encodeErr := customMarshaler.Marshal(odataFilters)
	if encodeErr == nil && len(values) > 0 {
		values = "?" + strings.ReplaceAll(values, "&", "&$")
	}
	return values, encodeErr
}
