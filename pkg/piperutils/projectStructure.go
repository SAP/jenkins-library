package piperutils

import "path/filepath"

// ProjectStructure describes a directory containing source code
type ProjectStructure struct {
	directory string
}

// UsesMta returns `true` if the project structure directory contains typical files for mta projects (mta.yaml, mta.yml), `false` otherwise
func (projectStructure *ProjectStructure) UsesMta() bool {
	return projectStructure.anyFileExists("mta.yaml", "mta.yml")
}

// UsesMaven returns `true` if the project structure directory contains a pom.xml file, false otherwise
func (projectStructure *ProjectStructure) UsesMaven() bool {
	return projectStructure.anyFileExists("pom.xml")
}

// UsesNpm returns `true` if the project structure directory contains a package.json file, false otherwise
func (projectStructure *ProjectStructure) UsesNpm() bool {
	return projectStructure.anyFileExists("package.json")
}

func (projectStructure *ProjectStructure) anyFileExists(candidates ...string) bool {
	for i := 0; i < len(candidates); i++ {
		exists, err := FileExists(filepath.Join(projectStructure.directory, candidates[i]))
		if err != nil {
			continue
		}
		if exists {
			return true
		}
	}
	return false
}
