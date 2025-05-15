package cnbutils

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const DefaultOrderPath = "/cnb/order.toml"

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

func loadExistingOrder(utils BuildUtils) (Order, error) {
	order := Order{
		Utils: utils,
	}

	orderReader, err := utils.Open(DefaultOrderPath)
	if err != nil {
		return Order{}, err
	}
	defer orderReader.Close()

	_, err = toml.NewDecoder(orderReader).Decode(&order)
	if err != nil {
		return Order{}, err
	}

	return order, nil
}

func newOrder(bpacks []string, utils BuildUtils) (Order, error) {
	buildpacksMeta, err := GetMetadata(bpacks, utils)
	if err != nil {
		return Order{}, err
	}

	return Order{
		Utils: utils,
		Order: []OrderEntry{{
			Group: buildpacksMeta,
		}},
	}, nil
}

func CreateOrder(bpacks, preBpacks, postBpacks []string, dockerCreds string, utils BuildUtils) (Order, error) {
	if dockerCreds != "" {
		os.Setenv("DOCKER_CONFIG", filepath.Dir(dockerCreds))
	}

	var order Order
	var err error
	if len(bpacks) == 0 {
		order, err = loadExistingOrder(utils)
		if err != nil {
			return Order{}, err
		}
	} else {
		order, err = newOrder(bpacks, utils)
		if err != nil {
			return Order{}, err
		}
	}

	for idx := range order.Order {
		preMetadata, err := GetMetadata(preBpacks, utils)
		if err != nil {
			return Order{}, err
		}

		postMetadata, err := GetMetadata(postBpacks, utils)
		if err != nil {
			return Order{}, err
		}

		order.Order[idx].Group = append(preMetadata, order.Order[idx].Group...)
		order.Order[idx].Group = append(order.Order[idx].Group, postMetadata...)
	}

	return order, nil
}
