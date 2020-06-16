// Package rackfile includes structs and functions to setup and retrieve information from rackfile.yamls
package rackfile

import (
	"errors"
	"io/ioutil"
	"log"
	"os"

	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/fileutil"

	"gopkg.in/yaml.v2"
)

// RackFile struct that will define the yaml structure of the rackfile.yml
type RackFile struct {
	Plugins []string
	Themes  []string
}

// UnmarshalRackFile unmarshals a rackfile from a given yaml path
func UnmarshalRackFile(yamlPath string) (*RackFile, error) {
	data, err := ioutil.ReadFile(yamlPath) //nolint, as the file is only being unmarshaled
	if err != nil {
		return nil, err
	}

	file := &RackFile{}

	err = yaml.Unmarshal(data, file)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// MarshalRackFile will Marshal a given RackFile struct to yaml data.
func (r RackFile) MarshalRackFile() (*[]byte, error) {
	data, err := yaml.Marshal(r)
	if err != nil {
		return nil, err
	}

	return &data, err
}

// CreateExampleRackFile creates a example Rackfile for the user so he can see how plugins and themes can be integrated.
func CreateExampleRackFile() error {
	plugins := []string{
		"PluginExample",
		"PluginLatestExample:latest",
		"PluginVersionExample:1.0.0",
	}

	themes := []string{
		"ThemeExample",
		"ThemeLatestExample:latest",
		"ThemeVersionExample:1.0.0",
	}

	rackfile := RackFile{plugins, themes}

	data, err := yaml.Marshal(&rackfile)
	if err != nil {
		log.Fatalf("Marshal - error: %v\n", err)
		return err
	}

	filename := "rackfile.yaml"

	exists, ok := fileutil.ObjectExists(filename)
	if ok != nil {
		log.Fatalf("objectExists - error: %v\n", ok)
		return ok
	}

	if exists {
		log.Fatalln("rackfile already exists - Aborting")
		return errors.New("rackfile already exists")
	}

	err = ioutil.WriteFile(filename, data, os.ModePerm)
	if err != nil {
		log.Fatalf("WriteFile - error: %v\n", err)
	}

	return err
}
