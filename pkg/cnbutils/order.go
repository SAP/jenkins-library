package cnbutils

import (
	"bytes"

	"github.com/pelletier/go-toml"
)

func (o Order) Save(path string) error {
	var buf bytes.Buffer

	err := toml.NewEncoder(&buf).Encode(o)
	if err != nil {
		return err
	}

	err = o.Futils.FileWrite(path, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
