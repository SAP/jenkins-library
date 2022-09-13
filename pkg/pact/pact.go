package pact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/SAP/jenkins-library/pkg/log"
)

type Utils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	command.ExecRunner
	piperhttp.Sender
	piperutils.FileUtils
	GetExitCode() int
}

type utilsBundle struct {
	*command.Command
	*piperhttp.Client
	*piperutils.Files
}

func NewUtilsBundle() Utils {
	utils := utilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Client:  &piperhttp.Client{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

// PactSpec represents an AsyncPact.json file
type PactSpec struct {
	Consumer Consumer `json:"consumer"`
	Provider Provider `json:"provider"`
}

// Consumer represents the consumer of the given contract
type Consumer struct {
	Name string `json:"name"`
}

// Provider represents the provider of the given contract
type Provider struct {
	Name string `json:"name"`
}

// LatestPactsForProviderTagResp represents a response from the pact broker which contains url link(s) to
// the pact contracts associated with a specific provider
type LatestPactsForProviderTagResp struct {
	Links Links `json:"_links"`
}

// Links represents a slice of link structures
type Links struct {
	PBPacts []Link `json:"pb:pacts"`
}

// Link represents a single link to a contract that exists in the pact-broker
type Link struct {
	HRef  string `json:"href"`
	Title string `json:"title"`
	Name  string `json:"name"`
}

// PactBrokerClient represents a connection to the pact-broker
type PactBrokerClient struct {
	hostname   string
	brokerUser string
	brokerPass string
}

// ErrNotFound is an error message that will be returned when no contracts have been published for associated provider
var ErrNotFound = fmt.Errorf("404: no consumer tests found for provider")

// NewPactBrokerClient initializes and returns a PactBrokerClient with the values passed in as arguments
func NewPactBrokerClient(hostname, user, pass string) *PactBrokerClient {
	return &PactBrokerClient{
		hostname:   hostname,
		brokerUser: user,
		brokerPass: pass,
	}
}

// LatestPactsForProviderByTag retrieves and returns links to pact contracts associated with provider and tag passed in as arguments.
func (pc *PactBrokerClient) LatestPactsForProviderByTag(provider, tag string, utils Utils) (*LatestPactsForProviderTagResp, error) {

	url := fmt.Sprintf("https://%s:%s@%s/pacts/provider/%s/latest/%s",
		pc.brokerUser,
		pc.brokerPass,
		pc.hostname,
		provider,
		tag)
	
	resp, err := sendRequest(http.MethodGet, url, nil, utils)
	if err != nil {
		if err == ErrNotFound {
			log.Entry().Infof("No consumer tests found for provider: %s", provider)
		}
		return nil, err
	}

	pactLinks := &LatestPactsForProviderTagResp{}
	if err := json.Unmarshal(resp, pactLinks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return pactLinks, nil
}

// DownloadPactContract will send a GET request to the pact broker for a specific pact contract using the url passed in as an argument.
// It return the response and any http error if encountered.
func (pc *PactBrokerClient) DownloadPactContract(url string, utils Utils) ([]byte, error) {
	s := strings.Split(url, "https://")
	formattedURL := fmt.Sprintf("https://%s:%s@%s", pc.brokerUser, pc.brokerPass, s[1])
	return sendRequest(http.MethodGet, formattedURL, nil, utils)
}

// PublishPact executes the pact publish cli tool to upload contract to pact broker
// It returns an error if any are encountered.
func (pc *PactBrokerClient) PublishPact(cfg *PublishConfig, pactContract string, utils Utils, stdout io.Writer) error {
	log.Entry().Infof("Consumer pact version: %s", cfg.GitCommit)
	log.Entry().Infof("Tag: %s", cfg.GitSourceBranch)
	log.Entry().Infof("Pact file: %s", pactContract)

	// Find executable for pact cli tool
	pactPublishExecutable, err := utils.LookPath("pact")
	if err != nil {
		return fmt.Errorf("failed to find pact executable: %w", err)
	}

	// Parameters for pact cli tool
	args := []string{
		pactPublish,
		fmt.Sprintf("--broker-username=%s", pc.brokerUser),
		fmt.Sprintf("--broker-password=%s", pc.brokerPass),
		fmt.Sprintf("--broker-base-url=https://%s", pc.hostname),
		fmt.Sprintf("--consumer-app-version=%s", cfg.GitCommit),
		pactContract,
		fmt.Sprintf("--tag=%s", cfg.GitSourceBranch),
	}

	var pactLog bytes.Buffer
	utils.Stdout(&pactLog)
	err = utils.RunExecutable(pactPublishExecutable, args...)
	utils.Stdout(stdout)
	log.Entry().Print(pactLog)
	if err != nil {
		log.Entry().WithError(err).Error("Error running command %v", pactPublishExecutable)
		if strings.Contains(pactLog.String(), "Each pact must be published with a unique consumer version number.") {
			log.Entry().Warning("Consumer version already published to broker. No change will be made. This could result from re-triggering a pipeline on the same commit ID.")
			return nil
		}
		return err
	}

	// Contract succesfully published to pact broker
	return nil
}