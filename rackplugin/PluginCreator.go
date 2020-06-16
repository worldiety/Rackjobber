package rackplugin

import (
	"log"
	"os"
	"path/filepath"

	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/fileutil"
)

// createPluginSkeleton creates a plugin skeleton containing the base .php, plugin.xml and rackspec
func createPluginSkeleton(pluginName string, opts map[string]string) {
	createPluginFolder(pluginName)
	createBasePHP(pluginName)
	createPluginXML(pluginName, opts)
}

// createPluginFolder creates a folder for the new plugin
func createPluginFolder(pluginName string) {
	currentDir, _ := os.Getwd()
	pluginDir := filepath.Join(currentDir, pluginName)

	err := fileutil.CreateDirIfNotExistant(pluginDir)
	if err != nil {
		log.Fatalf("CreateDirIfNotExistant - error: %v\n", err)
	}
}

func createBasePHP(pluginName string) {
	currentDir, _ := os.Getwd()
	pluginDir := filepath.Join(currentDir, pluginName)

	basePhp := "<?php\n\n"
	basePhp += "namespace " + pluginName + ";\n\n"
	basePhp += "use Shopware\\Components\\Plugin;\n\n"
	basePhp += "class " + pluginName + " extends Plugin\n"
	basePhp += "{\n\n}\n"

	data := []byte(basePhp)

	err := fileutil.CreateOrWriteFile(filepath.Join(pluginDir, pluginName+".php"), data)
	if err != nil {
		log.Fatalf("CreateOrWriteFile - error: %v\n", err)
	}
}

func createPluginXML(pluginName string, opts map[string]string) {
	currentDir, _ := os.Getwd()
	pluginDir := filepath.Join(currentDir, pluginName)

	pluginXML := "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n"
	pluginXML += "<plugin>\n"
	pluginXML += "	<label>" + pluginName + "</label>\n\n"

	if len(opts["description"]) > 0 {
		pluginXML += "<description>" + opts["description"] + "</description>\n"
	}

	if len(opts["version"]) > 0 {
		pluginXML += "	<version>" + opts["version"] + "</version>\n\n"
	}

	if len(opts["copyright"]) > 0 {
		pluginXML += "<copyright>" + opts["copyright"] + "</copyright>\n"
	}

	if len(opts["license"]) > 0 {
		pluginXML += "<license>" + opts["license"] + "</license>\n"
	}

	if len(opts["link"]) > 0 {
		pluginXML += "<link>" + opts["link"] + "</link>\n"
	}

	if len(opts["author"]) > 0 {
		pluginXML += "<author>" + opts["author"] + "</author>\n"
	}

	if len(opts["minVersion"]) > 0 || len(opts["maxVersion"]) > 0 {
		pluginXML += "<compatibility"
		if len(opts["minVersion"]) > 0 {
			pluginXML += " minVersion=\"" + opts["minVersion"] + "\""
		}

		if len(opts["maxVersion"]) > 0 {
			pluginXML += " maxVersion=\"" + opts["maxVersion"] + "\""
		}

		pluginXML += "/>\n"
	}

	if len(opts["version"]) > 0 {
		pluginXML += "	<changelog version=\"" + opts["version"] + "\">\n"
	} else {
		pluginXML += "	<changelog version=\"0.0.1\">\n"
	}

	pluginXML += "		<changes>Initial version</changes>\n"
	pluginXML += "	</changelog>\n"
	pluginXML += "</plugin>"

	data := []byte(pluginXML)

	err := fileutil.CreateOrWriteFile(filepath.Join(pluginDir, "plugin.xml"), data)
	if err != nil {
		log.Fatalf("CreateOrWriteFile - error: %v\n", err)
	}
}
