package sonar

import (
	"github.com/SAP/jenkins-library/pkg/log"
	sonarAPI "github.com/magicsong/sonargo/sonar"
	"github.com/pkg/errors"
)

func (api *IssueService) createClient() error {
	log.Entry().Debug("creating new api client for '%s'", api.Host)
	client, err := sonarAPI.NewClient(api.Host, api.Token, "")
	if err != nil {
		return errors.Wrap(err, "failed to connect to Sonar server")
	}
	api.client = client
	return nil
}
