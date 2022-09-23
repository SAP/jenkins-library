package pact

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

// Config represents all configuration options used as flags for publish and verify commands
type PublishConfig struct {
	PathToPactsFolder         string
	PactBrokerBaseURL         string
	PactBrokerUsername        string
	PactBrokerPassword        string
	PactBrokerToken           string
	OrgOrigin                 string
	OrgAlias                  string
	GitPullID                 string
	BuildID                   string
	GitTargetBranch           string
	GitRepo                   string
	GitSourceBranch           string
	GitCommit                 string
	GitProvider               string
	EnforceOpenAPIValidation  bool
	EnforceAsyncAPIValidation bool
	SystemNamespace           string
	Utils                     Utils
	StdOut                    io.Writer
}

func (cfg *PublishConfig) Report() *ReportData {
	return &ReportData{
		OrgOrigin:   cfg.OrgOrigin,
		OrgAlias:    cfg.OrgAlias,
		GitProvider: cfg.GitProvider,
		GitRepo:     cfg.GitRepo,
		GitCommit:   cfg.GitCommit,
		GitPullID:   cfg.GitPullID,
		BuildID:     cfg.BuildID,
		GitBranch:   cfg.GitSourceBranch,
	}
}

// command passed to pact cli tool
const pactPublish = "publish"

// ExecPactPublish will publish all pact files found in pathToPactsFolder to the pactBroker and upload number of contracts published to ci report server
func (p *PublishConfig) ExecPactPublish() error {
	log.Entry().Info("Executing HTTP Pact Verify")

	reportData := p.Report()
	pactClient := NewPactBrokerClient(p.PactBrokerBaseURL, p.PactBrokerUsername, p.PactBrokerPassword)

	// Ensures the path to the pact files is in the correct format
	p.PathToPactsFolder = EnsureValidDir(p.PathToPactsFolder)

	// Open directory that contains pact contracts to be published
	pactFiles, err := p.Utils.Glob(p.PathToPactsFolder + "**")
	if len(pactFiles) == 0 || err != nil {
		return fmt.Errorf("no pact files found in: '%s'; if this is unexpected please check your configuration", p.PathToPactsFolder)
	}

	// numOfContractsWritten is value sent in ci report
	numOfContractsWritten := len(pactFiles)
	log.Entry().Infof("Publishing all json files in %s to the Pact Broker", p.PathToPactsFolder)
	log.Entry().Infof("Number of Contracts to publish %d", numOfContractsWritten)
	for _, pactFile := range pactFiles {
		log.Entry().Infof("Path to pact contract: %s%s", p.PathToPactsFolder, pactFile)
		pactContractSpec := &PactSpec{}
		if err = ReadAndUnmarshalFile(pactFile, pactContractSpec, p.Utils); err != nil {
			return fmt.Errorf("failed to parse pact file: %w", err)
		}
		consumer := pactContractSpec.Consumer.Name
		provider := pactContractSpec.Provider.Name
		// Enforce Naming conventions for provider and consumer as specified in contract
		if ok := enforceNaming(p.GitRepo, consumer, provider); !ok {
			return fmt.Errorf("pact contract does not follow the correct naming conventions: %s", pactFile)
		}

		// Publish pact to brokeer
		if err := pactClient.PublishPact(p, pactFile, p.Utils, p.StdOut); err != nil {
			return fmt.Errorf("failed publishing to server: %w", err)
		}

	}

	report := Report{}
	filePath := "pactPublishReport.json"
	if err := report.SaveReport(reportData, filePath, "Consumer", "consumer", strconv.Itoa(numOfContractsWritten), p.Utils); err != nil {
		return fmt.Errorf("error saving report: %w", err)
	}

	return nil
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
		return fmt.Errorf("failed to find pact executable 'pact': %w", err)
	}

	// Parameters for pact cli tool
	args := []string{
		pactPublish,
		pactContract,
		fmt.Sprintf("--broker-username=%s", pc.brokerUser),
		fmt.Sprintf("--broker-password=%s", pc.brokerPass),
		fmt.Sprintf("--broker-base-url=https://%s", pc.hostname),
		fmt.Sprintf("--consumer-app-version=%s", cfg.GitCommit),
		fmt.Sprintf("--tag=%s", cfg.GitSourceBranch),
	}

	var pactLog bytes.Buffer
	utils.Stdout(&pactLog)
	err = utils.RunExecutable(pactPublishExecutable, args...)
	utils.Stdout(stdout)
	log.Entry().Print(pactLog)
	if err != nil {
		log.Entry().WithError(err).Errorf("Error running command %v", pactPublishExecutable)
		if strings.Contains(pactLog.String(), "Each pact must be published with a unique consumer version number.") {
			log.Entry().Warning("Consumer version already published to broker. No change will be made. This could result from re-triggering a pipeline on the same commit ID.")
			return nil
		}
		return err
	}

	// Contract succesfully published to pact broker
	return nil
}

// enforceNaming enforces naming conventions for the consumer & provider specified in the pact contract.
// It returns a boolean value representing the enforcement status.
func enforceNaming(gitRepo, consumerName, providerName string) bool {
	if consumerName != fmt.Sprintf("%s-async", gitRepo) && consumerName != fmt.Sprintf("%s-http", gitRepo) {
		log.Entry().Errorf("Consumer name is NOT using the correct naming strategy: %s. Use either %s-async or %s-http.", consumerName, gitRepo, gitRepo)
		return false
	}

	if !strings.HasSuffix(providerName, "-async") && !strings.HasSuffix(providerName, "-http") {
		log.Entry().Errorf("Provider name is not using the correct naming strategy: %s. Use either providerName-async or providerName-http.", providerName)
		return false
	}

	return true
}
