// Package rackshopstore includes structs and operations to setup and manage shops for rackjobber to work on
package rackshopstore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/fileutil"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackshop"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackssh"
)

func getShopStorePath() (*string, error) {
	resPath, err := fileutil.GetAppFolderPath()
	if err != nil {
		return nil, err
	}

	storePath := filepath.Join(*resPath, "shopstore.yaml")

	return &storePath, nil
}

func getShopStore() (*ShopStore, error) {
	storePath, err := getShopStorePath()
	if err != nil {
		return nil, err
	}

	exists, err := fileutil.ObjectExists(*storePath)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("shopstore.yaml does not exist")
	}

	store, err := UnmarshalShopStore(*storePath)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func readOrCreateShopStore() (*ShopStore, error) {
	storePath, err := getShopStorePath()
	if err != nil {
		return nil, err
	}

	exists, err := fileutil.ObjectExists(*storePath)
	if err != nil {
		return nil, err
	}

	var shopStore *ShopStore

	if exists {
		shopStore, err = UnmarshalShopStore(*storePath)
		if err != nil {
			return nil, err
		}
	} else {
		shopStore = &ShopStore{}
	}

	return shopStore, nil
}

func appendShopToStore(shop rackshop.RackShop) error {
	shopStore, err := readOrCreateShopStore()
	if err != nil {
		return err
	}

	shopStore.Shops = append(shopStore.Shops, shop)

	data, err := shopStore.MarshalShopStore()
	if err != nil {
		return err
	}

	storePath, err := getShopStorePath()
	if err != nil {
		return err
	}

	err = fileutil.CreateOrWriteFile(*storePath, *data)
	if err != nil {
		return err
	}

	return nil
}

func proveRemoteShopConnection(shop rackshop.RackShop) bool {
	return rackssh.CheckConnection(shop.Address, shop.GetRemoteConfig())
}

// AddShopWithFile will add a shop based on a file where the informations are available
func AddShopWithFile(path string) error {
	rackShop, err := rackshop.UnmarshalRackShop(path)
	if err != nil {
		return err
	}

	connected := proveRemoteShopConnection(*rackShop)

	if !connected {
		fmt.Println("Could not connect to shop, shop will not be added to store")
		return errors.New("unable to connect to remote shop")
	}

	err = appendShopToStore(*rackShop)
	if err != nil {
		return err
	}

	return nil
}

// AddShop will add a shop to rackjobber based on the passed flags
func AddShop(name string, address string, user string, password string, sdir string, container string) error {
	rackShop := rackshop.RackShop{
		Name:        name,
		Address:     address,
		User:        user,
		Password:    password,
		ShopwareDir: sdir,
		Container:   container,
	}

	connected := proveRemoteShopConnection(rackShop)

	if !connected {
		fmt.Println("Could not connect to shop, shop will not be added to store")
		return errors.New("unable to connect to remote shop")
	}

	err := appendShopToStore(rackShop)
	if err != nil {
		return err
	}

	return nil
}

// GetShopFromStore will return a shop config from the store.
func GetShopFromStore(name string) (*rackshop.RackShop, error) {
	shopStore, err := getShopStore()
	if err != nil {
		return nil, err
	}

	return shopStore.GetShopForName(name)
}

// RemoveShopFromStore will remove a shop with a given name from the shop store
func RemoveShopFromStore(name string) error {
	shopStore, err := getShopStore()
	if err != nil {
		return err
	}

	indexToRemove := -1

	for index, shop := range shopStore.Shops {
		if strings.Compare(shop.Name, name) == 0 {
			indexToRemove = index
		}
	}

	if indexToRemove == -1 {
		fmt.Printf("Could not find shop with name: %v\n", name)
		return errors.New("shop not found")
	}

	shops := shopStore.Shops

	shops[indexToRemove] = shops[len(shops)-1]
	shopStore.Shops = shops[:len(shops)-1]

	storePath, err := getShopStorePath()
	if err != nil {
		return nil
	}

	if len(shopStore.Shops) == 0 {
		return os.Remove(*storePath)
	}

	data, err := shopStore.MarshalShopStore()
	if err != nil {
		return nil
	}

	return fileutil.CreateOrWriteFile(*storePath, *data)
}

// ListShopsFromStore returns all shops that are currently configured in the shop store
func ListShopsFromStore() (*[]rackshop.RackShop, error) {
	shopStore, err := getShopStore()
	if err != nil {
		return nil, err
	}

	return &shopStore.Shops, nil
}
