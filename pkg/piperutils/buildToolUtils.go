package piperutils

import "path/filepath"

type buildTools struct {
	directory string
}

// UsesMta returns `true` if the cwd contains typical files for mta projects (mta.yaml, mta.yml), `false` otherwise
func (b *buildTools) UsesMta() bool {
	return b.anyFileExists("mta.yaml", "mta.yml")
}

// UsesMaven returns `true` if the cwd contains a pom.xml file, false otherwise
func (b *buildTools) UsesMaven() bool {
	return b.anyFileExists("pom.xml")
}

// UsesNpm returns `true` if the cwd contains a package.json file, false otherwise
func (b *buildTools) UsesNpm() bool {
	return b.anyFileExists("package.json")
}

func (b *buildTools) anyFileExists(candidates ...string) bool {
	for i := 0; i < len(candidates); i++ {
		exists, err := FileExists(filepath.Join(b.directory, candidates[i]))
		if err != nil {
			return false
		}
		if exists {
			return true
		}
	}
	return false
}
