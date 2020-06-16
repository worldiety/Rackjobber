// Package rackspec includes the struct and functions to setup and manage the rackspec.yaml
package rackspec

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/fileutil"

	"gopkg.in/yaml.v2"
)

// RackSpec Model that will be used by go-yaml to create the yaml data.
type RackSpec struct {
	Name          string        `yaml:"name"`
	Version       string        `yaml:"version"`
	Description   string        `yaml:"description"`
	Author        string        `yaml:"author"`
	Compatibility Compatibility `yaml:"compatibility"`
	Source        Source        `yaml:"source"`
	Theme         string        `yaml:"theme"`
}

//Source Interface with clone URL
type Source struct {
	GIT string `yaml:"GIT"`
}

// Compatibility contains the compatibility information for a plugin
type Compatibility struct {
	MinVersion string `xml:"minVersion,attr"`
	MaxVersion string `xml:"maxVersion,attr"`
}

// UnmarshalRackSpec will unmarshal a Rackspec from a given path
func UnmarshalRackSpec(yamlPath string) (*RackSpec, error) {
	data, err := ioutil.ReadFile(yamlPath) //nolint, as the file is only being unmarshaled
	if err != nil {
		return nil, err
	}

	spec := &RackSpec{}

	err = yaml.Unmarshal(data, spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

// MarshalRackSpec will marshal an existing RackSpec struct to yaml data.
func (r RackSpec) MarshalRackSpec() (*[]byte, error) {
	data, err := yaml.Marshal(r)
	if err != nil {
		return nil, err
	}

	return &data, err
}

// CreateDefaultRackSpec will create a template rackspeck.yml file, that can be used for a plugin or theme.
// The default yml-File is for a plugin.
func CreateDefaultRackSpec(name string) error {
	defaultSourceURL := "REPO_URL"
	defaultSource := Source{defaultSourceURL}

	defaultSpec := RackSpec{
		Name:          name,
		Version:       "0.0.1",
		Description:   "Add a short description here",
		Author:        "Authorname",
		Compatibility: Compatibility{"5.5", "5.5"},
		Source:        defaultSource,
		Theme:         "",
	}

	data, err := yaml.Marshal(&defaultSpec)
	if err != nil {
		log.Fatalf("yaml.Marshal - error: %v\n", err)
		return err
	}

	directory := filepath.Join(filepath.Join(".", string(filepath.Separator)), "rackspec")

	exists, ok := fileutil.ObjectExists(directory)

	if ok != nil {
		log.Fatalf("dirExists(directory) - error: %v\n", ok)
		return ok
	}

	if !exists {
		err := os.MkdirAll(directory, os.ModePerm)
		if err != nil {
			log.Fatalf("MkdirAll(directory) - error: %v\n", err)
		}
	}

	path := filepath.Join(directory, fmt.Sprintf("%s.yaml", name))

	exists, ok = fileutil.ObjectExists(path)
	if ok != nil {
		log.Fatalf("dirExists(path) - error: %v\n", ok)
		return ok
	}

	if exists {
		log.Fatalf("%v already exists! Aborting\n", path)
		return errors.New("Template file exists")
	}

	err = ioutil.WriteFile(path, data, os.ModePerm)
	if err != nil {
		log.Fatalf("WriteFile - error: %v\n", err)
	}

	return err
}

// FindRackSpecInCurrentDir will try to find the rackspec file in the current directory
func FindRackSpecInCurrentDir() (*string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return FindRackSpecInDir(dir)
}

// FindRackSpecInDir will try to find the rackspec file in the passed directory
func FindRackSpecInDir(directory string) (*string, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.Contains(file.Name(), "rackspec") {
			specPath := filepath.Join(directory, file.Name())
			return &specPath, nil
		}
	}

	return nil, errors.New("RackSpec not found")
}
