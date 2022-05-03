package piperenv

type Artifact struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path"`
}

type Artifacts []Artifact

func (a Artifacts) FindByName(name string) Artifacts {
	var filtered Artifacts

	for _, artifact := range a {
		if artifact.Name == name {
			filtered = append(filtered, artifact)
		}
	}
	return filtered
}
