package config

const (
	gcpJsonKeyFilePath = "gcpJsonKeyFilePath"
	gcsFolderPath      = "gcsFolderPath"
	gcsBucketID        = "gcsBucketId"
)

var (
	reportingFilter = []string{
		gcpJsonKeyFilePath,
		gcsFolderPath,
		gcsBucketID,
	}
)

func (s *StepConfig) mixinReportingConfig(configs ...map[string]interface{}) {
	for _, config := range configs {
		s.mixIn(config, reportingFilter)
	}
}

// TODO: add aliases resolving for gcsBucketId (alias: pipelineId) and gcpJsonKeyFilePath (alias: jsonKeyFilePath)

// TODO: add commonPipelineEnvironment resolving for gcsFolderPath (env: gcsFolderPath)

// TODO: add vault resolving for gcpJsonKeyFilePath:
//              - $(vaultPath)/cumulus
//              - $(vaultBasePath)/$(vaultPipelineName)/cumulus
//              - $(vaultBasePath)/GROUP-SECRETS/cumulus)
