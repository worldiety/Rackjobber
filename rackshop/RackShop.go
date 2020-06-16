// Package rackshop includes the struct and functions to manage rackshops used in the rackshopstore
package rackshop

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

// RackShop struct that holds the information for a Shop that will be stored in the ShopStore
type RackShop struct {
	Name        string
	Address     string
	User        string
	Password    string
	ShopwareDir string
	Container   string
}

// UnmarshalRackShop will unmarshal a yaml file at a specified path.
// Returns a RackShop struct if the operation was successful
func UnmarshalRackShop(yamlPath string) (*RackShop, error) {
	data, err := ioutil.ReadFile(yamlPath) //nolint, only being unmarshaled
	if err != nil {
		return nil, err
	}

	shop := &RackShop{}

	err = yaml.Unmarshal(data, shop)
	if err != nil {
		return nil, err
	}

	return shop, nil
}

// MarshalRackShop will marshal the current RackShop struct and return yaml data.
func (r RackShop) MarshalRackShop() (*[]byte, error) {
	data, err := yaml.Marshal(r)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// GetRemoteConfig will return a remote config, with that a ssh connection to the shop should be possible
func (r RackShop) GetRemoteConfig() *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: r.User,
		Auth: []ssh.AuthMethod{
			getPrivateKeyFile(),
		},
		HostKeyCallback: ssh.FixedHostKey(getHostKey(r.Address)),
	}
}

//getPrivateKeyFile retrieves the executing device's rsa key for authentication
func getPrivateKeyFile() ssh.AuthMethod {
	key, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"))
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
		fmt.Println("Please setup the initial ssh connection manually")
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
		fmt.Println("Please setup the initial ssh connection manually")
	}

	return ssh.PublicKeys(signer)
}

//getHostKey retrieves the host's public Key for authentication from the executing device's .ssh/known_hosts
func getHostKey(host string) ssh.PublicKey {
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		log.Fatalf("known_hosts not found: %v\n", err)
		fmt.Println("Please setup the initial ssh connection manually")
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatalf("file.Close - error: %v\n", err)
		}
	}()

	hostFileRowLength := 3

	var hostKey ssh.PublicKey

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != hostFileRowLength {
			continue
		}

		if strings.Contains(fields[0], host) {
			var err error

			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				log.Fatalf("error parsing %q: %v", fields[2], err)
			}

			break
		}
	}

	if hostKey == nil {
		log.Fatalf("no hostkey found for %s", host)
		fmt.Println("Please setup the initial ssh connection manually")
	}

	return hostKey
}
