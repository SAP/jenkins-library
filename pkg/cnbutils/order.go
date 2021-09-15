package cnbutils

import (
	"bytes"

	"github.com/pelletier/go-toml"
)

type Order struct {
	Order []OrderEntry `toml:"order"`
	Utils BuildUtils   `toml:"-"`
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
	var buf bytes.Buffer

	err := toml.NewEncoder(&buf).Encode(o)
	if err != nil {
		return err
	}

	err = o.Utils.FileWrite(path, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
