package piperenv

type Artifact struct {
	Kind      string `json:"kind,omitempty"`
	LocalPath string `json:"localPath,omitempty"`
	//ToDo might want to introduce a remote path (to point to where the artifact is pushed after build)
	Name string `json:"name,omitempty"`
}

type Artifacts []Artifact

func (a Artifacts) FindByKind(kind string) Artifacts {
	var filtered Artifacts

	for _, artifact := range a {
		if artifact.Kind == kind {
			filtered = append(filtered, artifact)
		}
	}

	return filtered
}

func (a Artifacts) FindByName(name string) Artifacts {
	var filtered Artifacts

	for _, artifact := range a {
		if artifact.Name == name {
			filtered = append(filtered, artifact)
		}
	}

	return filtered
}
