package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/fileutil"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackcommands"
)

func main() {
	var app = cli.NewApp()

	setup()
	info(app)
	commands(app)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func setup() {
	rackresourcePath, _ := fileutil.GetAppFolderPath()
	_ = fileutil.CreateDirIfNotExistant(*rackresourcePath)
	_ = fileutil.CreateDirIfNotExistant(filepath.Join(*rackresourcePath, "repos"))
}

func info(app *cli.App) {
	app.Name = "Rackjobber"
	app.Usage = "Command line tool to deploy plugins and themes to your shopware installation"
	authors := []*cli.Author{
		{
			Name:  "Marcel Tuchner",
			Email: "marcel.tuchner@worldiety.com",
		},
		{
			Name:  "Thomas Riedel",
			Email: "thomas.riedel@worldiety.com",
		},
		{
			Name:  "Jonas Hodde",
			Email: "jonas.hodde@worldiety.com",
		},
	}
	app.Authors = authors
	app.Version = "0.0.1"
}

func commands(app *cli.App) {
	app.Commands = []*cli.Command{
		rackcommands.SetupCommand(),
		rackcommands.AccountCommand(),
		rackcommands.RepoCommand(),
		rackcommands.ShopCommand(),
		rackcommands.PluginCommand(),
		rackcommands.UpCommand(),
	}
}
