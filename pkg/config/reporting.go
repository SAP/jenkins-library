package config

type ReportingParams struct {
	Parameters []StepParameters
}

var ReportingParameters = ReportingParams{
	Parameters: []StepParameters{
		{
			Name:    "gcpJsonKeyFilePath",
			Aliases: []Alias{{Name: "jsonKeyFilePath"}},
			ResourceRef: []ResourceReference{
				{
					Name:    "gcpJsonKeyFilePathVaultSecretName",
					Type:    "vaultSecret",
					Default: "gcp",
				},
			},
		},
		{
			Name: "gcsFolderPath",
			ResourceRef: []ResourceReference{
				{
					Name:  "commonPipelineEnvironment",
					Param: "gcsFolderPath",
				},
			},
		},
		{
			Name:    "gcsBucketId",
			Aliases: []Alias{{Name: "pipelineId"}},
		},
	},
}

func (r ReportingParams) GetResourceParameters(path, name string) map[string]interface{} {
	resourceParams := map[string]interface{}{}

	for _, param := range r.Parameters {
		for _, res := range param.ResourceRef {
			if res.Name == name {
				if val := getParameterValue(path, res, param); val != nil {
					resourceParams[param.Name] = val
				}
			}
		}
	}
	return resourceParams
}

func (r ReportingParams) getStepFilters() StepFilters {
	var filters StepFilters
	reportingFilter := r.getReportingFilter()
	filters.All = append(filters.All, reportingFilter...)
	filters.General = append(filters.General, reportingFilter...)
	filters.Steps = append(filters.Steps, reportingFilter...)
	filters.Stages = append(filters.Stages, reportingFilter...)
	return filters
}

func (r ReportingParams) getReportingFilter() []string {
	var reportingFilter []string
	for _, param := range r.Parameters {
		reportingFilter = append(reportingFilter, param.Name)
	}
	return reportingFilter
}

func (s *StepConfig) mixinReportingConfig(configs ...map[string]interface{}) {
	reportingFilter := ReportingParameters.getReportingFilter()
	for _, config := range configs {
		s.mixIn(config, reportingFilter)
	}
}
