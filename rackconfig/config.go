// Package rackconfig includes the struct and functions to setup and retrieve information from the config
package rackconfig

import (
	"encoding/hex"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"

	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/fileutil"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackinput"
)

//Config contains configuration information for Rackjobber
type Config struct {
	GITAccounts      []GITAccount `yaml:"GIT"`
	MandatoryPlugins []string     `yaml:"plugins"`
}

//GITAccount contains GIT account information for a user
type GITAccount struct {
	Domain     string `yaml:"domain"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password,omitempty"`
	Inkeychain bool   `yaml:"inkeychain"`
}

//AddAccount adds or modifies an account
//Password will be hashed. Only implemented to prevent clearly visible passwords in the config. This is no encryption!
func AddAccount(domain string, username string, password string, fromkeychain bool) {
	appFolderPath, err := fileutil.GetAppFolderPath()
	if err != nil {
		log.Fatalf("Application directory could not be found: %v\n", err)
	}

	configPath := filepath.Join(*appFolderPath, "config.yaml")
	cfg := Config{}
	replaced := false

	if exists, _ := fileutil.ObjectExists(configPath); exists {
		cfg = GetConfig()
		for i, account := range cfg.GITAccounts {
			if account.Domain == domain && !replaced {
				replacemsg := "Account for domain " + account.Username + ":" + domain + " already exists. Replace? (y/n):"

				replace := rackinput.AwaitTextInput(replacemsg)
				if replace == "yes" || replace == "y" {
					cfg.GITAccounts[i] = GITAccount{domain, username, password, fromkeychain}
					replaced = true
				}
			}
		}
	}

	if !replaced {
		cfg.GITAccounts = append(cfg.GITAccounts, GITAccount{domain, username, password, fromkeychain})
	}

	configData, _ := yaml.Marshal(&cfg)

	err = fileutil.CreateOrWriteFile(configPath, configData)
	if err != nil {
		log.Fatalf("CreateOrWriteFile - error: %v\n", err)
	}
}

//RemoveAccount removes an existing account
func RemoveAccount(domain string) {
	appFolderPath, err := fileutil.GetAppFolderPath()
	if err != nil {
		log.Fatalf("Application directory could not be found: %v\n", err)
	}

	configPath := filepath.Join(*appFolderPath, "config.yaml")
	cfg := Config{}
	found := false

	if exists, _ := fileutil.ObjectExists(configPath); exists {
		cfg = GetConfig()
		for i, account := range cfg.GITAccounts {
			if account.Domain == domain && !found {
				tmpI, tmpLast := cfg.GITAccounts[i], cfg.GITAccounts[len(cfg.GITAccounts)-1]
				cfg.GITAccounts[len(cfg.GITAccounts)-1], cfg.GITAccounts[i] = tmpI, tmpLast
				cfg.GITAccounts = cfg.GITAccounts[:len(cfg.GITAccounts)-1]
				found = true

				fmt.Println("Account removed.")
			}
		}
	}

	if !found {
		fmt.Println("Account not found.")
	}

	configData, err := yaml.Marshal(&cfg)
	if err != nil {
		log.Fatalf("yaml.Marshal - error: %v\n", err)
	}

	err = fileutil.CreateOrWriteFile(configPath, configData)
	if err != nil {
		log.Fatalf("CreateOrWriteFile - error: %v\n", err)
	}
}

//GetConfig returns the current configuration
func GetConfig() Config {
	appFolderPath, err := fileutil.GetAppFolderPath()
	if err != nil {
		log.Fatalf("Application directory could not be found: %v\n", err)
	}

	configPath := filepath.Join(*appFolderPath, "config.yaml")

	config, _ := fileutil.ReadFile(configPath)
	if config != nil {
		cfg := &Config{}

		err = yaml.Unmarshal(*config, &cfg)
		if err != nil {
			log.Fatalf("yaml.Unmarshal - error: %v\n", err)
		}

		return *cfg
	}

	return Config{
		[]GITAccount{},
		[]string{},
	}
}

//GetGITAuth returns account data if an account has been set for given domain
func GetGITAuth(domain string) http.BasicAuth {
	passInKeyChain := false
	auth := http.BasicAuth{}

	config := GetConfig()
	for _, account := range config.GITAccounts {
		if account.Domain == domain {
			auth.Username = account.Username

			if strings.Contains(auth.Username, "@") {
				auth.Username = strings.Replace(auth.Username, "@", "%40", 1)
			}

			if account.Inkeychain {
				passInKeyChain = true
				break
			} //full auth stored in config

			password, _ := hex.DecodeString(account.Password)
			auth.Password = string(password)

			return auth
		}
	}

	if passInKeyChain { //only username stored in config, get password from Keychain
		u, p, err := getGitPassFromKeychain(auth.Username, domain)
		if err == nil {
			auth.Username = u
			auth.Password = p

			return auth
		}
	}

	for auth.Username == "" { //Neither Password, nor Username stored in config, setup new Account
		auth.Username = rackinput.AwaitTextInput("Username for domain " + domain + "?\n")
	}

	u, p, err := getGitPassFromKeychain(auth.Username, domain)
	if err == nil { //Full Authentication retrieved from Keychain
		auth.Username = u
		auth.Password = p
		AddAccount(domain, auth.Username, "", true)

		return auth
	}

	enc := ""
	for auth.Password == "" { //No Authentication from Keychain, manual password input required
		enc = rackinput.AwaitPasswordInput("Password for domain " + domain + "?\n")
		dec, _ := hex.DecodeString(enc)
		auth.Password = string(dec)
	}
	AddAccount(domain, auth.Username, enc, false)

	return auth
}
