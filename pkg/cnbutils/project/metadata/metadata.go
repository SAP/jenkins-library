package metadata

import "github.com/SAP/jenkins-library/pkg/git"

type project struct {
	Source source `toml:"source"`
}

type source struct {
	Type     string   `toml:"type"`
	Metadata metadata `toml:"metadata"`
	Version  version  `toml:"version"`
}

type metadata struct {
	Repository string `toml:"repository"`
	Revision   string `toml:"revision"`
}

type version struct {
	Commit string `toml:"commit"`
}

func WriteProjectMetadata(src string) error {
	_, err := git.PlainOpen(src)
	if err != nil {
		return err
	}

	return nil
}
