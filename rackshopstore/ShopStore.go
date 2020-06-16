package rackshopstore

import (
	"io/ioutil"
	"log"
	"strings"

	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackshop"
	"gopkg.in/yaml.v2"
)

// ShopStore struct that defines the structure for the shopStore.yaml
type ShopStore struct {
	Shops []rackshop.RackShop
}

// UnmarshalShopStore will unmarshal a yaml file at a specified path.
// Returns a ShopStore struct if the operation was successful
func UnmarshalShopStore(yamlPath string) (*ShopStore, error) {
	data, err := ioutil.ReadFile(yamlPath) //nolint, as the file is only being unmarshaled 
	if err != nil {
		return nil, err
	}

	store := &ShopStore{}

	err = yaml.Unmarshal(data, store)
	if err != nil {
		return nil, err
	}

	return store, nil
}

// MarshalShopStore will marshal the current ShopStore object and return yaml data if the operation was successful
func (s ShopStore) MarshalShopStore() (*[]byte, error) {
	data, err := yaml.Marshal(s)
	if err != nil {
		return nil, err
	}

	return &data, err
}

// GetShopForName will return the shop configuration for a specific name
func (s ShopStore) GetShopForName(name string) (*rackshop.RackShop, error) {
	for _, shop := range s.Shops {
		if strings.Compare(shop.Name, name) == 0 {
			return &shop, nil
		}
	}

	log.Fatalf("Shop with name %v not found.", name)

	return nil, nil
}
