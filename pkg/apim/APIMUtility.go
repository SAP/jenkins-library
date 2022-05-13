package apim

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/pkg/errors"
)

//Utils for apim
type Utils interface {
	InitAPIM() error
	IsPayloadJSON() bool
}

//Bundle struct
type Bundle struct {
	APIServiceKey, Host, Payload string
	Client                       piperhttp.Sender
}

//InitAPIM() fumnction initialize APIM bearer token for API access
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
		return errors.Wrap(tokenErr, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token.AccessToken)
	httpClient.SetOptions(clientOptions)
	return nil
}

//IsJSON checks given string is valid json or not
func (apim *Bundle) IsPayloadJSON() bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(apim.Payload), &js) == nil
}
