package sonar

import (
	sonarAPI "github.com/magicsong/sonargo/sonar"
	"github.com/pkg/errors"
)

func (api *IssueService) createClient() error {
	client, err := sonarAPI.NewClient(api.Host, api.Token, "")
	if err != nil {
		return errors.Wrap(err, "failed to connect to Sonar server")
	}
	api.client = client
	return nil
}
