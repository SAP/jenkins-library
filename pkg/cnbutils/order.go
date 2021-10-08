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
	Group []BuildPackMetadata `toml:"group" json:"group"`
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
