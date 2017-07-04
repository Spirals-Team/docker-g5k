package main

import (
	"os"

	"github.com/Spirals-Team/docker-g5k/command"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine/log"
)

var (
	// AppVersion stores the application version
	AppVersion = "head(git)"
	// appFlags stores the application global flags
	appFlags = []cli.Flag{}
	// cliCommands stores the application commands
	cliCommands = []cli.Command{command.CreateClusterCliCommand, command.ListClusterCliCommand, command.RemoveClusterCliCommand}
)

func main() {
	app := cli.NewApp()

	app.Version = AppVersion
	app.Flags = appFlags
	app.Commands = cliCommands

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}
}
