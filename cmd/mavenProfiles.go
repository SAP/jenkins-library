package cmd

// containsProfile reports whether target appears in the profiles slice.
func containsProfile(profiles []string, target string) bool {
	for _, p := range profiles {
		if p == target {
			return true
		}
	}
	return false
}

// buildEffectiveProfiles returns the profiles to pass to Maven: builtin guards
// (skipped when the caller already supplied them) followed by the caller's profiles.
// This keeps the step idempotent when the Groovy layer has already prepended the guards.
func buildEffectiveProfiles(userProfiles []string) []string {
	builtins := buildBuiltinProfiles(userProfiles)
	var result []string
	for _, p := range builtins {
		if !containsProfile(userProfiles, p) {
			result = append(result, p)
		}
	}
	return append(result, userProfiles...)
}
