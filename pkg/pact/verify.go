package pact

import (
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

// VerifyConfig represents all configuration options used in verify stage
type VerifyConfig struct {
	PathToAsyncFile           string
	PathToSwaggerFile         string
	PactBrokerBaseURL         string
	PactBrokerUsername        string
	PactBrokerPassword        string
	PactBrokerToken           string
	OrgOrigin                 string
	OrgAlias                  string
	GitProvider               string
	GitRepo                   string
	GitPullID                 string
	BuildID                   string
	GitTargetBranch           string
	GitSourceBranch           string
	GitCommit                 string
	EnforceOpenAPIValidation  bool
	EnforceAsyncAPIValidation bool
	SystemNamespace           string
	Provider                  string
	Utils                     Utils
}

func (v *VerifyConfig) Report() *ReportData {
	return &ReportData{
		OrgOrigin:   v.OrgOrigin,
		OrgAlias:    v.OrgAlias,
		GitProvider: v.GitProvider,
		GitRepo:     v.GitRepo,
		GitCommit:   v.GitCommit,
		GitPullID:   v.GitPullID,
		BuildID:     v.BuildID,
		GitBranch:   v.GitSourceBranch,
	}
}

// Values represent http & async contract test results that will be sent to CI Report server.
const (
	PASSED = "Passed"
	FAILED = "Failed"
	NA     = "N/A"
)

// pathToAsyncPactFolder represents the path to the folder where pact contracts which are associated with verifying provider will be downloaded to.
const pathToAsyncPactFolder = "./async_verify_pacts/"

type enforcementConfig struct {
	EnforceOpenAPIValidation  bool `json:"enforceOpenAPIValidation,omitempty"`
	EnforceAsyncAPIValidation bool `json:"enforceAsyncAPIValidation,omitempty"`
}

// ErrEnforcement error returned if threshold(s) are not met
var (
	ErrEnforcement = fmt.Errorf("pipeline enforcement failed")
)

// ExecPactVerify will execute applicable http and async contract tests and upload the results to the CI Report server.
func (v *VerifyConfig) ExecPactVerify() error {
	reportData := v.Report()

	// Removes suffix in case it was wrongly specified in the configuration
	// '-http' and '-async' suffix will be appended during each respective stage automatically
	v.Provider = strings.TrimSuffix(v.Provider, "-http")
	v.Provider = strings.TrimSuffix(v.Provider, "-async")

	log.Entry().Info("Executing HTTP Pact Verify")
	// If no consumer contracts have been written for provider result will be set to 0 to prevent pipeline failure
	httpExitCode, numOfHTTPContracts, httpErr := v.verifyHTTP()
	if httpErr != nil {
		// do not fail here, finalize testing first ...
		log.Entry().WithError(httpErr).Error("failed to verify HTTP Pact tests")
	}
	httpReportResult := reportStatus(httpExitCode, numOfHTTPContracts)
	log.Entry().Infof("HTTP Result: %v, HTTP Exit Code: %d, Number of tests: %d", httpReportResult, httpExitCode, numOfHTTPContracts)

	log.Entry().Info("Executing ASYNC Pact Verify")
	// If no consumer contracts have been written for provider result will be set to 0 to prevent pipeline failure
	asyncExitCode, numOfAsyncContracts, asynchErr := v.verifyAsync()
	if asynchErr != nil {
		log.Entry().WithError(asynchErr).Error("failed to verify Asynch Pact tests")
	}
	asyncReportResult := reportStatus(asyncExitCode, numOfAsyncContracts)
	log.Entry().Infof("ASYNC Result: %v, Async Exit Code: %d, Number of tests: %d", asyncReportResult, asyncExitCode, numOfAsyncContracts)

	// save report
	report := Report{}
	filePath := "pactVerifyReport.json"
	if err := report.SaveReport(reportData, filePath, "Provider", "provider", fmt.Sprintf("%s:%s", httpReportResult, asyncReportResult), v.Utils); err != nil {
		return fmt.Errorf("error saving report: %w", err)
	}

	// Fail pipeline if any contract tests failed
	if asyncExitCode != 0 || httpExitCode != 0 {
		return fmt.Errorf("contract tests failed, http: %v, asynch: %v", httpReportResult, asyncReportResult)
	}
	// Fail pipeline if verification failed
	if asynchErr != nil || httpErr != nil {
		return fmt.Errorf("contract tests validation failed, http: %v, asynch: %v", httpErr, asynchErr)
	}
	return nil
}

// verifyHTTP will verify http contracts for given provider using swagger-mock-validator.
// It return two ints and an error representing the validators exit code, the number of contract tests that were associated with given provider, and an error if encountered
func (v *VerifyConfig) verifyHTTP() (exitCode, numOfTests int, err error) {
	pactClient := NewPactBrokerClient(v.PactBrokerBaseURL, v.PactBrokerUsername, v.PactBrokerPassword)
	provider := fmt.Sprintf("%s-http", v.Provider)
	log.Entry().Infof("Validating provider %s Swagger against consumer contracts tagged '%s'", provider, v.GitTargetBranch)
	log.Entry().Infof("Path to swagger: %s", v.PathToSwaggerFile)
	// Downloads the links of contracts associated with provider. Link count is used to
	// to prevent enforcement from failing pipeline if no consumer tests were written.
	pactLinks, err := pactClient.LatestPactsForProviderByTag(provider, v.GitTargetBranch, v.Utils)

	if err == ErrNotFound {
		return exitCode, numOfTests, nil
	}

	if err != nil {
		return exitCode, numOfTests, err
	}

	numberOfTests := len(pactLinks.Links.PBPacts)
	if numberOfTests == 0 {
		return exitCode, numberOfTests, nil
	}

	if exists, err := v.Utils.FileExists(v.PathToSwaggerFile); !exists || err != nil {
		log.Entry().Infof("No swagger file for provider detected in: %s", v.PathToSwaggerFile)
		exitCode = 1
		return exitCode, numberOfTests, nil
	}

	// Find executable to swagger-mock-validator
	swaggerExecutable, err := v.Utils.LookPath("swagger-mock-validator")
	if err != nil {
		return 1, numberOfTests, err
	}

	// arguments passed to swagger-mock-validator tool
	args := []string{
		v.PathToSwaggerFile,
		fmt.Sprintf("https://%s", v.PactBrokerBaseURL),
		fmt.Sprintf("--provider=%s", provider),
		fmt.Sprintf("--tag=%s", v.GitTargetBranch),
		fmt.Sprintf("--user=%s:%s", v.PactBrokerUsername, v.PactBrokerPassword),
	}

	// Run swagger-mock-validator and return exit code if test does not pass to fail pipeline
	if err = v.Utils.RunExecutable(swaggerExecutable, args...); err != nil {
		return v.Utils.GetExitCode(), numberOfTests, err
	}

	// Contract test Passed
	return exitCode, numberOfTests, err
}

// verifyAsync will verify async contracts for given provider using async-validator.
// It return two ints and an error representing the validators exit code, the number of contract tests that were associated with given provider, and an error if encountered
func (v *VerifyConfig) verifyAsync() (exitCode, numOfTests int, err error) {
	log.Entry().Infof("Validating provider %s-async asyncapidoc against consumer contracts tagged '%s", v.Provider, v.GitTargetBranch)
	log.Entry().Infof("Path to async: %s", v.PathToAsyncFile)

	numberOfTests, err := v.downloadContractsToDisk()

	if err == ErrNotFound {
		return exitCode, numOfTests, nil
	}

	if err != nil {
		return exitCode, numOfTests, err
	}

	if numberOfTests == 0 {
		return exitCode, numberOfTests, nil
	}

	// Fail pipeline if no provider asyncapidoc.json file has been written
	if exists, err := v.Utils.FileExists(v.PathToAsyncFile); !exists || err != nil {
		log.Entry().Infof("No async file for provider detected in: %v", v.PathToAsyncFile)
		exitCode = 1
		return exitCode, numberOfTests, nil
	}

	// Find executable to async-api-validator
	asyncValidatorExecutable, err := v.Utils.LookPath("async-api-validator")
	if err != nil {
		exitCode = 1
		return exitCode, numberOfTests, err
	}

	// arguments passed to async-api-validator tool
	args := []string{
		fmt.Sprintf("--pathToPactFolder=%s", pathToAsyncPactFolder),
		fmt.Sprintf("--pathToAsyncFile=%s", v.PathToAsyncFile),
	}

	// Run async-api-validator and return exit code if contract test fails to fail pipeline
	if err = v.Utils.RunExecutable(asyncValidatorExecutable, args...); err != nil {
		return v.Utils.GetExitCode(), numberOfTests, err
	}

	// Contract test passed
	return exitCode, numberOfTests, err
}

// downloadContractsToDisk will download and save to disk all consumer contracts which are associated with the calling provider.
// It returns two values, an int representing the number of links retrieved, and error if encountered
func (v *VerifyConfig) downloadContractsToDisk() (int, error) {
	pactClient := NewPactBrokerClient(v.PactBrokerBaseURL, v.PactBrokerUsername, v.PactBrokerPassword)
	provider := fmt.Sprintf("%s-async", v.Provider)
	pactLinks, err := pactClient.LatestPactsForProviderByTag(provider, v.GitTargetBranch, v.Utils)
	if err != nil {
		return 0, err
	}
	numberOfTests := len(pactLinks.Links.PBPacts)
	log.Entry().Infof("%v consumer tests found", numberOfTests)
	if err := EnsureDir(pathToAsyncPactFolder, v.Utils); err != nil {
		return 0, fmt.Errorf("failed to ensure that directory is existing: %w", err)
	}
	log.Entry().Infof("Saving async pact contracts to: %s", pathToAsyncPactFolder)

	for _, link := range pactLinks.Links.PBPacts {
		resp, err := pactClient.DownloadPactContract(link.HRef, v.Utils)
		if err != nil {
			return 0, fmt.Errorf("failed to download contract: %w", err)
		}

		fileName := fmt.Sprintf("%s/%s-%s.json", pathToAsyncPactFolder, link.Name, provider)

		if err := v.Utils.WriteFile(fileName, resp, 0o777); err != nil {
			return 0, fmt.Errorf("failed to write pact file %v: %w", fileName, err)
		}
	}

	return numberOfTests, nil
}

// reportStatus accepts in as arguments an exit code and the number of tests.
// It returns a status of NA Passed or Failed based on the arguments passed in.
func reportStatus(validatorExitCode, numberOfTests int) string {
	status := NA
	if numberOfTests > 0 && validatorExitCode == 0 {
		status = PASSED
	} else if numberOfTests > 0 && validatorExitCode != 0 {
		status = FAILED
	}
	return status
}

// Enforce checks the enforcement status of the associated provider repo.
// It will return an error if the repo does not comply with the associated enforcement threshholds.
func (v *VerifyConfig) Enforce(httpExitCode, asyncExitCode int) error {
	if !v.EnforceAsyncAPIValidation && !v.EnforceOpenAPIValidation {
		log.Entry().Info("enforcement is not enabled")
		return nil
	}

	if err := checkThreshold(v.GitRepo, httpExitCode, v.EnforceOpenAPIValidation, "openapi validation result"); err != nil {
		return ErrEnforcement
	}

	if err := checkThreshold(v.GitRepo, asyncExitCode, v.EnforceAsyncAPIValidation, "asyncapi validation result"); err != nil {
		return ErrEnforcement
	}

	return nil
}

func checkThreshold(repoName string, actual int, active bool, metricName string) error {
	if active && actual != 0 { // compare exit code (actual)
		log.Entry().Errorf("Repository %s did not pass enforcement %q, exit code: %v", repoName, metricName, actual)
		return ErrEnforcement
	} else if active && actual == 0 {
		log.Entry().Infof("[ PASS ] Repository %s passes enforcement: %q", repoName, metricName)
	} else {
		log.Entry().Infof("[ PASS ] Repository %s does not have enforcement enabled for %q", repoName, metricName)
	}
	return nil
}
