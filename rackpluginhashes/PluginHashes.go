// Package rackpluginhashes includes structs and functions
package rackpluginhashes

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

//PluginHashes struct that will define the yaml structure of the pluginhashes.yaml
type PluginHashes struct {
	Hashes []Hash `yaml:"hashes"`
}

//Hash struct that defines the yaml structure of a single Plugin in the pluginhashes.yaml
type Hash struct {
	Name string `yaml:"name"`
	Hash string `yaml:"hash"`
}

//UnmarshalPluginHashes unmarshals a pluginhashes.yaml file into a PluginHashes object
func UnmarshalPluginHashes(yamlPath string) (*PluginHashes, error) {
	data, err := ioutil.ReadFile(yamlPath) //nolint, as the file is only being unmarshaled
	if err != nil {
		return nil, err
	}

	file := &PluginHashes{}

	err = yaml.Unmarshal(data, file)
	if err != nil {
		return nil, err
	}

	return file, nil
}

//MarshalPluginHashes marshals a PluginHashes object into a bytestream, for it to be stored as yaml
func (h *PluginHashes) MarshalPluginHashes() (*[]byte, error) {
	data, err := yaml.Marshal(h)
	if err != nil {
		return nil, err
	}

	return &data, err
}

//GetHash returns the hash of the latest commit of a given Plugin
func (h *PluginHashes) GetHash(pluginName string) (string, error) {
	for _, hash := range h.Hashes {
		if hash.Name == pluginName {
			return hash.Hash, nil
		}
	}

	return "", errors.New("No hash found for plugin: " + pluginName)
}

//SetOrUpdateHash sets the Hash for a given Plugin. If the Plugin does not exist yet, it is created and added
func (h *PluginHashes) SetOrUpdateHash(pluginName, hashIn string) {
	for i, hash := range h.Hashes {
		if hash.Name == pluginName {
			h.Hashes[i].Hash = hashIn
			return
		}
	}

	h.Hashes = append(h.Hashes, Hash{pluginName, hashIn})
}

//MakeHashes returns a PluginHashes object containing the given Values
func MakeHashes(pluginName, hashIn string) *PluginHashes {
	return &PluginHashes{
		Hashes: []Hash{{Name: pluginName, Hash: hashIn}},
	}
}
