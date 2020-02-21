package piperutils

// UsesMta returns `true` if the cwd contains typical files for mta projects (mta.yaml, mta.yml), `false` otherwise
func UsesMta() bool {
	var mtaYaml, mtaYml bool
	var err error
	mtaYaml, err = FileExists("mta.yaml")
	if err != nil {
		// no action
	}
	mtaYml, err = FileExists("mta.yml")
	if err != nil {
		// no action
	}
	return mtaYaml || mtaYml
}

// UsesMaven returns `true` if the cwd contains a pom.xml file, false otherwise
func UsesMaven() bool {
	pom, err := FileExists("pom.xml")
	if err != nil {
		return false
	}
	return pom
}

// UsesNpm returns `true` if the cwd contains a package.json file, false otherwise
func UsesNpm() bool {
	packageJSON, err := FileExists("package.json")
	if err != nil {
		return false
	}
	return packageJSON
}
