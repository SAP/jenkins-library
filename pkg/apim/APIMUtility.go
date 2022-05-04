package apim

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/pkg/errors"
)

//APIMCommonUtils for apim
type APIMUtils interface {
	NewAPIM() error
}

//APIMBundle struct
type APIMBundle struct {
	APIServiceKey, Host string
	Client              piperhttp.Sender
}

func (apim *APIMBundle) NewAPIM() error {
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
