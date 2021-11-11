package buildsettings

type BuildSettingsInfo struct {
	Profiles                    []string `json:"profiles,omitempty"`
	Publish                     bool     `json:"publish,omitempty"`
	CreateBOM                   bool     `json:"createBOM,omitempty"`
	LogSuccessfulMavenTransfers bool     `json:"logSuccessfulMavenTransfers,omitempty"`
	GlobalSettingsFile          string   `json:"globalSettingsFile,omitempty"`
}

type BuildSettings struct {
	MavenBuild  []BuildSettingsInfo `json:"mavenBuild,omitempty"`
	NpmBuild    []BuildSettingsInfo `json:"npmBuild,omitempty"`
	DockerBuild []BuildSettingsInfo `json:"npmBuild,omitempty"`
	MtaBuild    []BuildSettingsInfo `json:"npmBuild,omitempty"`
}
