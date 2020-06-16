// Package rackup includes the implementation of the Up Command to update Plugins on a given shop
package rackup

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"gopkg.in/yaml.v2"

	sshrw "github.com/mosolovsa/go_cat_sshfilerw"

	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/fileutil"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/gitutil"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackconfig"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackinput"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackpluginhashes"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackshop"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackshopstore"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackspec"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackssh"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/repository"
)

//RackFile Interface to save read data
type RackFile struct {
	Plugins []string `yaml:"Plugins"`
}

//Up deploys plugins, that are listed in the Rackfile to a given shop
func Up(shopName string) {
	err := repository.UpdateRepos()
	if err != nil && err.Error() != "already up-to-date" {
		fmt.Printf("failed to update Master Repo: %v\n", err)
	} else {
		fmt.Println("Repository is up-to-date")
	}

	shop, _ := rackshopstore.GetShopFromStore(shopName)
	rackfile := getRackFile(shop)
	hashfile := getRackPluginHashes(shop)
	installedPlugins := getInstalledPlugins(shop)
	wantedPlugins := getMandatoryPlugins()

	if rackfile != nil {
	OUTER:
		for _, rackplugin := range rackfile.Plugins {
			for _, mandPlugin := range wantedPlugins {
				if mandPlugin == rackplugin {
					continue OUTER
				}
			}
			wantedPlugins = append(wantedPlugins, rackplugin)
		}
	}

	unwantedPlugins := getUnwantedPlugins(installedPlugins, wantedPlugins)
	for _, unwantedPlugin := range unwantedPlugins {
		if unwantedPlugin != "" {
			deletePlugin(unwantedPlugin, shop)
		}
	}

	initializeTheme(shop)
	updatePlugins(wantedPlugins, shop, hashfile)
	clearShopCache(shop)
	fmt.Println("Process finished.")
}

//updatePlugins iterates over given plugins and updates these remotely on given RackShop.
//It also checks for neseccity of updates for Plugins and updates the given Hashfile's Hashes
func updatePlugins(pluginsToUpdate []string, shop *rackshop.RackShop, shopHashes *rackpluginhashes.PluginHashes) {
	fmt.Println("updating plugins")

	for _, plugin := range pluginsToUpdate {
		pluginName := strings.Split(plugin, ":")[0]
		pluginRepo := getPluginRepo(pluginName)
		pluginVersion := getPluginVersion(plugin)
		pluginFlag := getPluginFlag(plugin)
		rackresourcesPath, _ := fileutil.GetAppFolderPath()
		rackspecPath := filepath.Join(*rackresourcesPath, "repos", pluginRepo, pluginName, pluginVersion)

		rackspec := getRackSpec(rackspecPath)
		if rackspec == nil {
			if reinstallMaster(pluginName) {
				return
			}

			continue
		}

		gitHash, err := gitutil.GetHashOfLastCommit(rackspec.Source.GIT, pluginVersion)
		if err != nil {
			log.Printf("Failed to retrieve hash for Plugin %v: %v", pluginName, err)
		}

		var isuptodate bool

		shopHashes, isuptodate = pluginIsUpToDate(pluginIsUpToDateArgs{pluginName, *gitHash, shopHashes})
		if isuptodate {
			continue
		}

		fmt.Println("Loading " + pluginName + " " + pluginVersion)
		rackssh.CloneGitToRemoteShop(shop, pluginName, rackspec.Source.GIT, pluginVersion)
		updatePlugin(pluginName, shop, pluginVersion)

		switch pluginFlag {
		case "noinstall":
		case "noactivate":
			installPlugin(pluginName, shop)
		case "nosettheme":
			installPlugin(pluginName, shop)
			activatePlugin(pluginName, shop)
		case "":
			installPlugin(pluginName, shop)
			activatePlugin(pluginName, shop)

			themeName := rackspec.Theme
			if len(themeName) != 0 {
				setTheme(themeName, shop)
			}
		}
	}

	fmt.Println("Finished updating plugins")

	err := updatePluginHashesToShop(shopHashes, shop)
	if err != nil {
		log.Fatalf("updatePluginHashesToShop - error: %v\n", err)
	}
}

func reinstallMaster(pluginName string) bool {
	in := ""
	for in != "y" && in != "n" {
		msg := "Plugin " + pluginName + " could not be found in current repos. Do you wish to reinstall your repo? (y/n)"
		in = rackinput.AwaitTextInput(msg)
	}

	if in == "y" {
		err := gitutil.ReinstallMaster()
		if err != nil {
			log.Fatalf("Faied to reinstall Master: %v\n", err)
		}

		fmt.Printf("successfully reinstalled Master Repository. Please restart rackjobber\n")

		return true
	}

	fmt.Println("Skipping this plugin")

	return false
}

type pluginIsUpToDateArgs struct {
	pluginName, gitHash string
	shopHashes          *rackpluginhashes.PluginHashes
}

//pluginIsUpToDate checks for last recent commit hash on shop and on gitlab
//compares these and updates the one on the shop.
//returns true if the plugin is up-to-date, otherwise returns false
func pluginIsUpToDate(args pluginIsUpToDateArgs) (*rackpluginhashes.PluginHashes, bool) {
	if args.shopHashes != nil {
		hashOnShop, err := args.shopHashes.GetHash(args.pluginName)
		if err != nil {
			fmt.Printf("no Hash found on shop for Plugin: %v\n", args.pluginName)
			args.shopHashes.SetOrUpdateHash(args.pluginName, args.gitHash)
		} else {
			if hashOnShop != args.gitHash {
				fmt.Println("Plugin is not up-to-date, updating plugin and hash")
				args.shopHashes.SetOrUpdateHash(args.pluginName, args.gitHash)
			} else {
				fmt.Printf("Plugin %v is up-to-date, skipping this one\n", args.pluginName)
				return args.shopHashes, true
			}
		}
	} else {
		log.Println("No Hashfile found on server")
		args.shopHashes = rackpluginhashes.MakeHashes(args.pluginName, args.gitHash)
	}

	return args.shopHashes, false
}

//getInstalledPlugins returns a list of all installed plugins
func getInstalledPlugins(shop *rackshop.RackShop) []string {
	return rackssh.GetRemoteDirsFromShop(shop, "custom/plugins")
}

// getMandatoryPlugins returns a list of plugins, that have to be installed first
func getMandatoryPlugins() []string {
	config := rackconfig.GetConfig()
	return config.MandatoryPlugins
}

//getUnwantedPlugins returns a list of installed plugins, that should be deleted
func getUnwantedPlugins(installedPlugins []string, wantedPlugins []string) []string {
	unwantedPlugins := make([]string, len(installedPlugins))

	for i, installedPlugin := range installedPlugins {
		unwantedPlugins[i] = installedPlugin

		for _, wantedPlugin := range wantedPlugins {
			if installedPlugin == strings.Split(wantedPlugin, ":")[0] {
				unwantedPlugins[i] = ""
			}
		}
	}

	return unwantedPlugins
}

//getRackFile returns the RackFile at given path
func getRackFile(shop *rackshop.RackShop) *RackFile {
	rackfile, err := rackssh.GetRemoteFileFromShop(shop, "custom/rackfile.yaml")
	if err != nil {
		log.Println("failed to get rackfile from shop")
		return nil
	}

	rf := &RackFile{}

	err = yaml.Unmarshal(rackfile, &rf)
	if err != nil {
		log.Printf("Unmarshal RackFile failed %v\n", err)
	}

	return rf
}

//getRackPluginHashes returns the PluginHashes at given path
func getRackPluginHashes(shop *rackshop.RackShop) *rackpluginhashes.PluginHashes {
	pluginhashes, err := rackssh.GetRemoteFileFromShop(shop, "custom/rackpluginhashes.yaml")
	if err != nil {
		log.Println("failed to get rackhashes from shop")
		return nil
	}

	ph := &rackpluginhashes.PluginHashes{}

	err = yaml.Unmarshal(pluginhashes, &ph)
	if err != nil {
		log.Printf("Unmarshal PluginHashes failed %v\n", err)
	}

	return ph
}

//updatePluginHashesToShop marshals the given PluginHashes into a yaml and uploads it to the rackshop
func updatePluginHashesToShop(ph *rackpluginhashes.PluginHashes, shop *rackshop.RackShop) error {
	hashfile, err := ph.MarshalPluginHashes()
	if err != nil {
		log.Printf("Marshal PluginHashes failed %v\n", err)
		return err
	}

	c, err := sshrw.NewSSHclt(shop.Address+":22", shop.GetRemoteConfig())
	if err != nil {
		log.Printf("Failed to connect to shop %v\n", err)
		return err
	}

	fileDir := filepath.Join(shop.ShopwareDir, "custom/rackpluginhashes.yaml")

	err = c.WriteFile(bytes.NewReader(*hashfile), fileDir)
	if err != nil {
		log.Printf("WriteFile - error: %v\n", err)
		return err
	}

	return nil
}

//getRackSpec returns the RackSpec at given path
func getRackSpec(path string) *rackspec.RackSpec {
	files, _ := ioutil.ReadDir(path)
	for _, file := range files {
		if strings.Contains(file.Name(), "rackspec.yaml") {
			rackspecData, err := ioutil.ReadFile(filepath.Join(path, file.Name())) //nolint, only being unmarshaled
			if err != nil {
				log.Printf("Failed reading RackSpec %v\n", err)
			}

			rs := &rackspec.RackSpec{}

			err = yaml.Unmarshal(rackspecData, &rs)
			if err != nil {
				log.Printf("Unmarshal RackSpec failed %v\n", err)
			}

			return rs
		}
	}

	return nil
}

//getPluginRepo returns the name of the repo a given plugin is from
func getPluginRepo(pluginName string) string {
	binPath, _ := fileutil.GetAppFolderPath()
	reposPath := filepath.Join(*binPath, "repos")

	repoDirs, err := ioutil.ReadDir(reposPath)
	if err != nil {
		log.Fatalf("ReadDir - error: %v\n", err)
	}

	for _, repoDir := range repoDirs {
		if repoDir.IsDir() {
			pluginPath := filepath.Join(reposPath, repoDir.Name(), pluginName)
			pluginInThisRepo, _ := exists(pluginPath)

			if pluginInThisRepo {
				return repoDir.Name()
			}
		}
	}

	return ""
}

//getPluginVersion returns the version of given plugin
func getPluginVersion(plugin string) string {
	split := strings.Split(plugin, ":")
	if len(split) > 1 {
		if split[1] == "latest" || split[1] == "" {
			return getLatestVersion(split[0])
		}

		return split[1]
	}

	return getLatestVersion(split[0])
}

//getLatestVersion returns the latest version of given plugin
func getLatestVersion(pluginName string) string {
	binPath, _ := fileutil.GetAppFolderPath()
	reposPath := filepath.Join(*binPath, "repos")

	repoDirs, err := ioutil.ReadDir(reposPath)
	if err != nil {
		log.Fatalf("ReadDir - error: %v\n", err)
	}

	latestVersion, _ := version.NewVersion("0.0.1")

	for _, repoDir := range repoDirs {
		if repoDir.IsDir() {
			pluginPath := filepath.Join(reposPath, repoDir.Name(), pluginName)
			pluginInThisRepo, _ := exists(pluginPath)

			if pluginInThisRepo {
				versionDirs, err := ioutil.ReadDir(pluginPath)
				if err != nil {
					log.Fatalf("ReadDir - error: %v\n", err)
				}

				for _, versionDir := range versionDirs {
					if versionDir.IsDir() {
						tempVersion, _ := version.NewVersion(versionDir.Name())
						if tempVersion.GreaterThan(latestVersion) {
							latestVersion = tempVersion
						}
					}
				}

				return latestVersion.String()
			}
		}
	}

	return ""
}

//getPluginFlag returns a plugins flag as defined in the RackFile
func getPluginFlag(plugin string) string {
	const pluginFlagPosition = 3

	split := strings.Split(plugin, ":")
	if len(split) == pluginFlagPosition {
		return split[2]
	}

	return ""
}

//deletePlugin deactivates, uninstalls and deletes a plugin from Shopware
func deletePlugin(pluginName string, shop *rackshop.RackShop) {
	fmt.Println("Deleting plugin " + pluginName + ".")
	commands := []string{
		"docker exec -i " + shop.Container + " php /var/www/html/bin/console sw:plugin:refresh -q",
		"docker exec -i " + shop.Container + " php /var/www/html/bin/console sw:plugin:deactivate -q " + pluginName,
		"docker exec -i " + shop.Container + " php /var/www/html/bin/console sw:plugin:uninstall -S -q " + pluginName,
		"docker exec -i " + shop.Container + " php /var/www/html/bin/console sw:plugin:delete -q " + pluginName,
	}

	for _, command := range commands {
		rackssh.RunRemoteCommandInShop(command, shop)
	}
}

//update Plugin fetches and checks out the Branch of the specified version and updates the Plugin on the given shop
func updatePlugin(pluginName string, shop *rackshop.RackShop, version string) {
	fmt.Println("Updating plugin " + pluginName + ".")
	pluginPath := filepath.Join(shop.ShopwareDir, "custom", "plugins", pluginName)
	refspec := "\"+refs/tags/" + version + ":refs/tags/" + version + "\""
	commands := []string{
		"docker exec -i " + shop.Container + " php /var/www/html/bin/console sw:plugin:refresh -q",
		"docker exec -i " + shop.Container + " php /var/www/html/bin/console sw:plugin:deactivate -q " + pluginName,
		"git -C " + pluginPath + " config remote.origin.fetch " + refspec,
		"git -C " + pluginPath + " fetch",
		"git -C " + pluginPath + " checkout " + version,
		"docker exec -i " + shop.Container + " php /var/www/html/bin/console sw:plugin:update -q " + pluginName,
	}

	for _, command := range commands {
		rackssh.RunRemoteCommandInShop(command, shop)
	}
}

//installPlugin installs a plugin
func installPlugin(pluginName string, shop *rackshop.RackShop) {
	fmt.Println("Installing plugin " + pluginName + ".")
	commands := []string{
		"docker exec -i " + shop.Container + " php /var/www/html/bin/console sw:plugin:refresh -q ",
		"docker exec -i " + shop.Container + " php /var/www/html/bin/console sw:plugin:install -q " + pluginName,
	}

	for _, command := range commands {
		rackssh.RunRemoteCommandInShop(command, shop)
	}
}

//activatePlugin activates an installed plugin
func activatePlugin(pluginName string, shop *rackshop.RackShop) {
	fmt.Println("Activating plugin " + pluginName + ".")

	command := "docker exec -i " + shop.Container + " php /var/www/html/bin/console sw:plugin:activate -q " + pluginName
	rackssh.RunRemoteCommandInShop(command, shop)
}

// setTheme sets the given theme for a given shop
func setTheme(themeName string, shop *rackshop.RackShop) {
	escapedThemeName := escapeThemeName(themeName)
	fmt.Printf("Setting Theme %v\n", escapedThemeName)
	command := "docker exec -i " + shop.Container + " php /var/www/html/bin/console wdy:theme:set -q " + escapedThemeName
	rackssh.RunRemoteCommandInShop(command, shop)
}

//initializeTheme resets a shop's theme to the Responsive theme
func initializeTheme(shop *rackshop.RackShop) {
	command := "docker exec -i " + shop.Container + " php /var/www/html/bin/console sw:theme:initialize -q"
	rackssh.RunRemoteCommandInShop(command, shop)
}

//clearShopCache clears a shop's cache
func clearShopCache(shop *rackshop.RackShop) {
	fmt.Println("Clearing shop cache.")

	command := "docker exec -i " + shop.Container + " php /var/www/html/bin/console sw:cache:clear -q "
	rackssh.RunRemoteCommandInShop(command, shop)
}

//exists returns if a path exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func escapeThemeName(themeName string) string {
	escaped := strings.ReplaceAll(themeName, " ", "_")
	return escaped
}
