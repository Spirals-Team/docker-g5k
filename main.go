package main

import (
	"os"

	"github.com/Spirals-Team/docker-g5k/command"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine/log"
)

var (
	// AppFlags stores cli flags for common parameters
	appFlags    = []cli.Flag{}
	cliCommands = []cli.Command{command.CreateClusterCliCommand, command.RemoveClusterCliCommand}
)

func main() {
	app := cli.NewApp()

	app.Flags = appFlags
	app.Commands = cliCommands

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}
}
