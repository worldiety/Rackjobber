// Package rackcommands includes the commands the user may call, to be used by the cli.App
package rackcommands

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/urfave/cli"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackconfig"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackinput"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackplugin"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/racksetup"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackshopstore"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackup"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/repository"
)

// SetupCommand is used to setup the default rackspec Repository
func SetupCommand() *cli.Command {
	return &cli.Command{
		Name:    "setup",
		Aliases: []string{"s"},
		Usage:   "Setup Rackjobber with your Rackspec Repository",
		Action: func(c *cli.Context) error {
			err := racksetup.Setup()
			if err != nil {
				log.Fatalf("Setup - error: %v\n", err)
			}
			return err
		},
	}
}

// AccountCommand is used for account related operations
func AccountCommand() *cli.Command {
	return &cli.Command{
		Name:     "account",
		Category: "Account actions",
		Subcommands: []*cli.Command{
			accountAddSubcommand(),
			accountRemoveSubcommand(),
			accountListSubcommand(),
		},
	}
}

func accountAddSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Add a new GIT account",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "domain, d",
				Usage: "Domain the GIT account belongs to",
			},
			&cli.StringFlag{
				Name:  "username, u",
				Usage: "Username of the GIT account",
			},
			&cli.StringFlag{
				Name:  "password, p",
				Usage: "Password of the GIT account",
			},
		},
		Action: func(c *cli.Context) error {
			domain, username, password := c.String("domain"), c.String("username"), c.String("password")
			for len(domain) == 0 {
				domain = rackinput.AwaitTextInput("GIT domain:")
			}
			for len(username) == 0 {
				username = rackinput.AwaitTextInput("Username:")
			}
			for len(password) == 0 {
				password = rackinput.AwaitPasswordInput("Password:")
			}
			rackconfig.AddAccount(domain, username, password, false)
			return nil
		},
	}
}

func accountRemoveSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Remove an existing GIT account",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "domain, d",
				Usage: "Domain the GIT account belongs to",
			},
		},
		Action: func(c *cli.Context) error {
			domain := c.String("domain")
			for len(domain) == 0 {
				domain = rackinput.AwaitTextInput("GIT domain:")
			}
			rackconfig.RemoveAccount(domain)
			return nil
		},
	}
}

func accountListSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all existing GIT accounts",
		Action: func(c *cli.Context) error {
			config := rackconfig.GetConfig()
			for _, account := range config.GITAccounts {
				fmt.Println(" - Domain: " + account.Domain)
				fmt.Println("\tUsername: " + account.Username)
			}
			return nil
		},
	}
}

// RepoCommand is used for repository related operations
func RepoCommand() *cli.Command {
	return &cli.Command{
		Name:     "repo",
		Category: "Repository Actions",
		Subcommands: []*cli.Command{
			repoAddSubcommand(),
			repoRemoveSubcommand(),
			repoListSubcommand(),
			repoUpdateSubcommand(),
			repoPushSubcommand(),
		},
	}
}

func repoAddSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Adds a repository to rackjobber",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "repoName, rn",
				Usage: "Name of the Repository that should be added",
			},
			&cli.StringFlag{
				Name:  "repoSource, rs",
				Usage: "Source of the Repository that should be added",
			},
		},
		Action: func(c *cli.Context) error {
			nameExists, repoName := proveStringCLI(c, "repoName")
			SourceExists, repoSource := proveStringCLI(c, "repoSource")

			if !nameExists || !SourceExists {
				return errors.New("required flag not provided")
			}

			err := repository.AddRepo(repoName, repoSource)
			if err != nil {
				log.Fatalf("AddRepo - error: %v\n", err)
			}

			return err
		},
	}
}

func repoRemoveSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Removes a repository from rackjobber",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "repo, r",
				Usage: "Repository that should be removed",
			},
		},
		Action: func(c *cli.Context) error {
			exists, repo := proveStringCLI(c, "repo")

			if !exists {
				return errors.New("required flag not provided")
			}

			err := repository.RemoveRepo(repo)
			if err != nil {
				log.Fatalf("RemoveRepo - error: %v\n", err)
			}

			return err
		},
	}
}

func repoListSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Lists all repos that are connected to rackjobber",
		Action: func(c *cli.Context) error {
			repos, err := repository.ListRepos()
			if err != nil {
				log.Fatalf("ListRepos - error: %v\n", err)
				return err
			}

			for repo, url := range *repos {
				fmt.Printf(" - Repository: %v\n", repo)
				fmt.Printf("\tURL: %v\n", url)
			}
			return nil
		},
	}
}

func repoUpdateSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Updates all repos that are connected to rackjobber",
		Action: func(c *cli.Context) error {
			err := repository.UpdateRepos()
			if err != nil {
				log.Fatalf("UpdateRepos - error: %v\n", err)
			}
			return err
		},
	}
}

func repoPushSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "push",
		Usage: "Pushes the rackspec.yaml from the current Directory to a specific repo",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "repoName, rn",
				Usage: "Repository in which the spec should be pushed, Default is master",
			},
		},
		Action: func(c *cli.Context) error {
			exists, repoName := proveStringCLI(c, "repoName")
			if !exists {
				return errors.New("required Flag not set")
			}

			err := repository.PushSpecToRepo(repoName)
			if err != nil {
				log.Fatalf("PushSpecToRepo - error: %v\n", err)
			}

			return err
		},
	}
}

// ShopCommand is used for shop related operations
func ShopCommand() *cli.Command {
	return &cli.Command{
		Name:     "shop",
		Category: "Shopware Installation actions",
		Subcommands: []*cli.Command{
			shopAddSubcommand(),
			shopRemoveSubcommand(),
			shopListSubcommand(),
			shopInitSubcommand(),
			shopIntegrateSubcommand(),
			shopDeintegrateSubcommand(),
		},
	}
}

func shopAddSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Adds a shop to rackjobber for the deployment",
		Flags: shopAddFlags(),
		Action: func(c *cli.Context) error {
			filePath := c.String("file")

			if len(filePath) > 0 {
				return rackshopstore.AddShopWithFile(filePath)
			}

			name := c.String("shopName")
			address := c.String("address")
			sshUser := c.String("sshuser")
			password := c.String("password")
			sdir := c.String("shopwareDir")
			container := c.String("container")

			for len(name) == 0 {
				name = rackinput.AwaitTextInput("Name (must not be empty):")
			}
			for len(address) == 0 {
				address = rackinput.AwaitTextInput("Address (must not be empty):")
			}
			for len(sshUser) == 0 {
				sshUser = rackinput.AwaitTextInput("SSH user (must not be empty):")
			}
			for len(password) == 0 {
				password = rackinput.AwaitPasswordInput("Password:")
			}
			for len(sdir) == 0 {
				sdir = rackinput.AwaitTextInput("Shopware directory (must not be empty):")
			}
			for len(container) == 0 {
				container = rackinput.AwaitTextInput("Docker container (must not be empty):")
			}
			return rackshopstore.AddShop(name, address, sshUser, password, sdir, container)
		},
	}
}

func shopAddFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "shopName, sn",
			Usage: "Name of the shop that should be added",
		},
		&cli.StringFlag{
			Name:  "address, a",
			Usage: "Address of the shop that will be added",
		},
		&cli.StringFlag{
			Name:  "sshuser, u",
			Usage: "SSH User to connect to the address of the shop",
		},
		&cli.StringFlag{
			Name:  "password, p",
			Usage: "Password for the ssh user",
		},
		&cli.StringFlag{
			Name:  "file, f",
			Usage: "File that contains the informations to add a shop to rackjobber",
		},
		&cli.StringFlag{
			Name:  "container, c",
			Usage: "Name of the docker container",
		},
		&cli.StringFlag{
			Name:  "shopwareDir, sdir",
			Usage: "Shopware directory on the remote machine",
		},
	}
}

func shopRemoveSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Removes a shop from the shopStore",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "shopName, sn",
				Usage: "The Name of the shop, that should be removed",
			},
		},
		Action: func(c *cli.Context) error {
			name := c.String("shopName")

			if len(name) > 0 {
				return rackshopstore.RemoveShopFromStore(name)
			}

			fmt.Println("Required Flag is missing")
			return errors.New("missing Required Flag")
		},
	}
}

func shopListSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Lists all shops that are currently in the shop store",
		Action: func(c *cli.Context) error {
			shops, err := rackshopstore.ListShopsFromStore()
			if err != nil {
				return err
			}

			for _, shop := range *shops {
				fmt.Printf(" - Shop: %v\n", shop.Name)
				fmt.Printf("\tAddress: %v\n", shop.Address)
				fmt.Printf("\tUser: %v\n", shop.User)
				fmt.Printf("\tShopwareDir: %v\n", shop.ShopwareDir)
				fmt.Printf("\tContainer: %v\n", shop.Container)
			}

			return nil
		},
	}
}

func shopInitSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initializes a shopware installation, that is compatible with rackjobber.",
		Action: func(c *cli.Context) error {
			fmt.Println("Sorry, but not implemented yet.")
			return nil
		},
	}
}

func shopIntegrateSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "integrate",
		Usage: "Integrates rackjobber into an existing shopware installation.",
		Action: func(c *cli.Context) error {
			fmt.Println("Sorry, but not implemented yet.")
			return nil
		},
	}
}

func shopDeintegrateSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "deintegrate",
		Usage: "Deintegrates rackjobber from this shopware installation",
		Action: func(c *cli.Context) error {
			fmt.Println("Sorry, but not implemented yet!")
			return nil
		},
	}
}

// PluginCommand is used for plugin related operations
func PluginCommand() *cli.Command {
	return &cli.Command{
		Name:     "plugin",
		Category: "Plugin specific actions",
		Subcommands: []*cli.Command{
			pluginInitSubcommand(),
			pluginIntegrateSubcommand(),
			pluginDeintegrateSubcommand(),
			pluginListSubcommand(),
		},
	}
}

func pluginInitSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initializes a plugin template, that is compatible with rackjobber.",
		Flags: pluginInitFlags(),
		Action: func(c *cli.Context) error {
			pluginName := strings.ReplaceAll(c.String("pluginName"), " ", "")
			opts := make(map[string]string)
			opts["description"] = c.String("description")
			opts["version"] = c.String("version")
			opts["author"] = c.String("author")
			opts["minVersion"] = c.String("minVersion")
			opts["maxVersion"] = c.String("maxVersion")
			opts["copyright"] = c.String("copyright")
			opts["license"] = c.String("license")
			opts["link"] = c.String("link")

			if opts["version"] == "" {
				opts["version"] = "0.0.1"
			}

			for len(pluginName) == 0 {
				pluginName = strings.ReplaceAll(rackinput.AwaitTextInput("Plugin name (must not be empty):"), " ", "")
			}

			rackplugin.InitializePlugin(pluginName, opts)
			return nil
		},
	}
}

func pluginInitFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "pluginName, pn",
			Usage: "Name of the new plugin",
		},
		&cli.StringFlag{
			Name:  "description, d",
			Usage: "Short description of the new plugin",
		},
		&cli.StringFlag{
			Name:  "version, v",
			Usage: "Initial version of the plugin",
		},
		&cli.StringFlag{
			Name:  "author, a",
			Usage: "Name of the author",
		},
		&cli.StringFlag{
			Name:  "minVersion, min",
			Usage: "Oldest version the plugin is compatible with",
		},
		&cli.StringFlag{
			Name:  "maxVersion, max",
			Usage: "Latest version the plugin is compatible with",
		},
		&cli.StringFlag{
			Name:  "copyright, c",
			Usage: "Copyright holder of the plugin",
		},
		&cli.StringFlag{
			Name:  "license, l",
			Usage: "License the plugin is used with",
		},
		&cli.StringFlag{
			Name:  "link",
			Usage: "Link to website of the developer",
		},
	}
}

func pluginIntegrateSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "integrate",
		Usage: "Integrates rackjobber into an existing plugin.",
		Action: func(c *cli.Context) error {
			return rackplugin.Integrate()
		},
	}
}

func pluginDeintegrateSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "deintegrate",
		Usage: "Deintegrates rackjobber from this plugin",
		Action: func(c *cli.Context) error {
			return rackplugin.Deintegrate()
		},
	}
}

func pluginListSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all plugins",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "repoName, rn",
				Usage: "Name a specific repo to list its plugins",
			},
		},
		Action: func(c *cli.Context) error {
			var plugins []string
			if len(c.String("repoName")) == 0 {
				plugins = rackplugin.GetAllPlugins()
			} else {
				plugins = rackplugin.GetPluginsFromRepo(c.String("repoName"))
			}
			for _, plugin := range plugins {
				fmt.Println(" - " + plugin)
			}
			return nil
		},
	}
}

// UpCommand is used to start the updating process
func UpCommand() *cli.Command {
	return &cli.Command{
		Name:    "up",
		Aliases: []string{"u"},
		Usage:   "Update and install Plugins and themes that are referenced in the rackfile",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "shopName, sn",
				Usage: "The name of the shop, that the plugins shall be deployed to",
			},
		},
		Action: func(c *cli.Context) error {
			shop := c.String("shopName")

			for len(shop) == 0 {
				shop = rackinput.AwaitTextInput("Shop name (must not be empty):")
			}

			rackup.Up(shop)
			return nil
		},
	}
}

func proveStringCLI(c *cli.Context, key string) (bool, string) {
	value := c.String(key)
	if strings.Compare(value, "") == 0 {
		log.Fatalf("No value for key: %v\n", key)
		return false, value
	}

	return true, value
}
