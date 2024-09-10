package config

func (s *StepConfig) FetchSecretFromTrustEngine(parameters []StepParameters, configs ...map[string]interface{}) {
	for _, config := range configs {
		s.mixIn(config, vaultFilter, StepData{})
		// when an empty filter is returned we skip the mixin call since an empty filter will allow everything
		if referencesFilter := getFilterForResourceReferences(parameters); len(referencesFilter) > 0 {
			s.mixIn(config, referencesFilter, StepData{})
		}
	}
}
