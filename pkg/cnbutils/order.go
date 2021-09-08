package cnbutils

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pelletier/go-toml"
)

type Order struct {
	Order  []OrderEntry      `toml:"order"`
	Futils *piperutils.Files `toml:"-"`
}

type OrderEntry struct {
	Group []BuildpackRef `toml:"group" json:"group"`
}

type BuildpackRef struct {
	ID       string `toml:"id"`
	Version  string `toml:"version"`
	Optional bool   `toml:"optional,omitempty" json:"optional,omitempty" yaml:"optional,omitempty"`
}

func (o Order) Save(path string) error {
	f, err := o.Futils.FileOpen(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	err = toml.NewEncoder(f).Encode(o)
	if err != nil {
		return err
	}

	return nil
}
