package cnbutils

import "github.com/SAP/jenkins-library/pkg/piperutils"

type CnbFileUtils interface {
	piperutils.FileUtils
	TempDir(string, string) (string, error)
	RemoveAll(string) error
	FileRename(string, string) error
}

type BuildPackMetadata struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Version     string    `json:"version,omitempty"`
	Description string    `json:"description,omitempty"`
	Homepage    string    `json:"homepage,omitempty"`
	Keywords    []string  `json:"keywords,omitempty"`
	Licenses    []License `json:"licenses,omitempty"`
}

type License struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
}

type Order struct {
	Order  []OrderEntry `toml:"order"`
	Futils CnbFileUtils `toml:"-"`
}

type OrderEntry struct {
	Group []BuildpackRef `toml:"group" json:"group"`
}

type BuildpackRef struct {
	ID       string `toml:"id"`
	Version  string `toml:"version"`
	Optional bool   `toml:"optional,omitempty" json:"optional,omitempty" yaml:"optional,omitempty"`
}
