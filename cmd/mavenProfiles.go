package cmd

// containsProfile reports whether target appears in the profiles slice.
// Shared by mtaBuild and mavenBuild.
func containsProfile(profiles []string, target string) bool {
	for _, p := range profiles {
		if p == target {
			return true
		}
	}
	return false
}
