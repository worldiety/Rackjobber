// Package rackplugin includes the struct and functions to setup
// and retrieve information from the plugins rackspec
package rackplugin

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/fileutil"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/gitutil"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackspec"

	"gopkg.in/yaml.v2"
)

// Plugin struct provides a bare skeleton for a plugin devinition of shopware
type Plugin struct {
	XMLName       xml.Name `xml:"plugin"`
	Name          string
	Version       string                 `xml:"version"`
	Author        string                 `xml:"author"`
	Description   string                 `xml:"description"`
	Compatibility rackspec.Compatibility `xml:"compatibility"`
}

// Integrate will add a default rackspec to the plugin in the current directory.
func Integrate() error {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Getwd - error: %v\n", err)
		return err
	}

	files, err := ioutil.ReadDir(currentDir)
	if err != nil {
		log.Fatalf("ReadDir - error: %v\n", err)
		return err
	}

	for _, file := range files {
		if strings.Contains(file.Name(), "plugin.xml") {
			repo, err := gitutil.GetRepoFromLocalDir(currentDir)
			if err != nil {
				log.Fatalf("OpenCurrentDirGit - error: %v\n", err)
				return err
			}

			if repo == nil {
				log.Fatalln("Repository is nil, can not determine source of plugin")
				return errors.New("Unable to open Repository")
			}

			url, err := gitutil.GetURLForRepo(*repo)
			if err != nil {
				log.Fatalf("GetURLForRepo - error: %v\n", err)
				return err
			}

			xmlData, err := ioutil.ReadFile(file.Name())
			if err != nil {
				log.Fatalf("ReadFile - error: %v\n", err)
				return err
			}

			err = writeYaml(currentDir, *url, xmlData)
			if err != nil {
				return err
			}
		}
	}

	log.Fatalf("Could not find plugin.xml in directory: '%v' to integrate rackspec\n", currentDir)

	return errors.New("No plugin.xml found")
}

func writeYaml(currentDir, url string, xmlData []byte) error {
	plugin := &Plugin{}

	err := xml.Unmarshal(xmlData, &plugin)
	if err != nil {
		log.Fatalf("xml.Unmarshal - error: %v\n", err)
	}

	comp := rackspec.Compatibility{
		MinVersion: plugin.Compatibility.MinVersion,
		MaxVersion: plugin.Compatibility.MaxVersion,
	}

	dirs, _ := ioutil.ReadDir(currentDir)
	for _, dir := range dirs {
		if strings.Contains(dir.Name(), ".php") {
			plugin.Name = strings.Split(dir.Name(), ".php")[0]
		}
	}

	theme := setupTheme(currentDir)
	spec := rackspec.RackSpec{
		Name:          plugin.Name,
		Version:       plugin.Version,
		Description:   plugin.Description,
		Author:        plugin.Author,
		Compatibility: comp,
		Source: rackspec.Source{
			GIT: url,
		},
		Theme: theme,
	}

	specData, err := yaml.Marshal(&spec)
	if err != nil {
		log.Fatalf("Marshal - error: %v\n", err)
		return err
	}

	rackSpecName := fmt.Sprintf("%v_rackspec.yaml", plugin.Name)
	rackSpecPath := filepath.Join(currentDir, rackSpecName)

	exists, err := fileutil.ObjectExists(rackSpecPath)
	if err != nil {
		log.Fatalf("ObjectExists - error: %v\n", err)
		return err
	}

	if exists {
		log.Fatalf("%v already exists! Aborting\n", rackSpecPath)
		return errors.New("RackSpecFile Already Exists")
	}

	err = ioutil.WriteFile(rackSpecPath, specData, os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Printf("Created rackspec at path: '%v', configure it when necessary\n", currentDir)

	return nil
}

func setupTheme(currentDir string) string {
	theme := ""

	themes, _ := ioutil.ReadDir(filepath.Join(currentDir, "Resources", "Themes", "Frontend"))
	for _, themeDir := range themes {
		if themeDir.IsDir() {
			theme = themeDir.Name()
		}
	}

	return theme
}

// Deintegrate will remove the rackspec from the plugin in the current directory.
func Deintegrate() error {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Getwd - error: %v\n", err)
		return err
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("ReadDir - error: %v\n", err)
		return err
	}

	for _, file := range files {
		if strings.Contains(file.Name(), "rackspec.yaml") {
			filePath := filepath.Join(dir, file.Name())

			err = os.Remove(filePath)
			if err != nil {
				log.Fatalf("Remove - error: %v\n", err)
				return err
			}

			log.Printf("Successfully removed %v\n", file.Name())

			return nil
		}
	}

	log.Fatalln("No rackspec in this directory - Aborting")

	return errors.New("No rackspec found")
}

// InitializePlugin willl initialize a Plugin repository with a integrated rackspec file.
func InitializePlugin(pluginName string, opts map[string]string) {
	createPluginSkeleton(pluginName, opts)
}

// GetAllPlugins returns a list of all plugins contained in all current added repos
func GetAllPlugins() []string {
	appPath, _ := fileutil.GetAppFolderPath()
	reposPath := filepath.Join(*appPath, "repos")
	plugins := []string{}

	repos, _ := ioutil.ReadDir(reposPath)
	for _, repo := range repos {
		if repo.IsDir() {
			pluginDirs, _ := ioutil.ReadDir(filepath.Join(reposPath, repo.Name()))
			for _, plugin := range pluginDirs {
				if plugin.IsDir() && plugin.Name() != ".git" {
					plugins = append(plugins, plugin.Name())
				}
			}
		}
	}

	return plugins
}

// GetPluginsFromRepo returns a list of all plugins in a given repo
func GetPluginsFromRepo(repoName string) []string {
	appPath, _ := fileutil.GetAppFolderPath()
	pluginsPath := filepath.Join(*appPath, "repos", repoName)
	plugins := []string{}

	pluginDirs, _ := ioutil.ReadDir(pluginsPath)
	for _, plugin := range pluginDirs {
		if plugin.IsDir() && plugin.Name() != ".git" {
			plugins = append(plugins, plugin.Name())
		}
	}

	return plugins
}
