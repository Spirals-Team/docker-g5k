package main

import (
	"os"

	"github.com/Spirals-Team/docker-g5k/command"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
)

var (
	// AppFlags stores cli flags for common parameters
	AppFlags = []cli.Flag{}

	// CliCommands stores cli commands and their flags
	CliCommands = []cli.Command{
		{
			Name:   "create-cluster",
			Usage:  "Create a new Docker Swarm cluster on Grid5000",
			Action: command.RunCreateClusterCommand,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "g5k-username",
					Usage: "Your Grid5000 account username",
					Value: "",
				},

				cli.StringFlag{
					Name:  "g5k-password",
					Usage: "Your Grid5000 account password",
					Value: "",
				},

				cli.StringFlag{
					Name:  "g5k-walltime",
					Usage: "Machine's lifetime (HH:MM:SS)",
					Value: "1:00:00",
				},

				cli.StringFlag{
					Name:  "g5k-ssh-private-key",
					Usage: "Path of your ssh private key",
					Value: mcnutils.GetHomeDir() + "/.ssh/id_rsa",
				},

				cli.StringFlag{
					Name:  "g5k-ssh-public-key",
					Usage: "Path of your ssh public key (default: \"<g5k-ssh-private-key>.pub\")",
					Value: "",
				},

				cli.StringFlag{
					Name:  "g5k-image",
					Usage: "Name of the image to deploy",
					Value: "jessie-x64-min",
				},

				cli.StringFlag{
					Name:  "g5k-resource-properties",
					Usage: "Resource selection with OAR properties (SQL format)",
					Value: "",
				},

				cli.StringFlag{
					Name:  "swarm-discovery",
					Usage: "Discovery service to use with Swarm",
					Value: "",
				},

				cli.StringFlag{
					Name:  "swarm-image",
					Usage: "Specify Docker image to use for Swarm",
					Value: "swarm:latest",
				},

				cli.StringFlag{
					Name:  "swarm-strategy",
					Usage: "Define a default scheduling strategy for Swarm",
					Value: "spread",
				},

				cli.StringSliceFlag{
					Name:  "swarm-opt",
					Usage: "Define arbitrary flags for Swarm master (can be provided multiple times)",
					Value: nil,
				},

				cli.StringSliceFlag{
					Name:  "swarm-join-opt",
					Usage: "Define arbitrary flags for Swarm join (can be provided multiple times)",
					Value: nil,
				},

				cli.BoolFlag{
					Name:  "swarm-master-join",
					Usage: "Make Swarm master join the Swarm pool",
				},

				cli.BoolFlag{
					Name:  "weave-networking",
					Usage: "Use Weave for networking",
				},

				cli.StringSliceFlag{
					Name:  "g5k-reserve-nodes",
					Usage: "Reserve nodes on a site (ex: lille:24)",
				},
			},
		},
		{
			Name:   "remove-cluster",
			Usage:  "Remove a Docker Swarm cluster from Grid5000",
			Action: command.RunRemoveClusterCommand,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "g5k-job-id",
					Usage: "Only remove nodes related to the provided job ID (By default ALL nodes from ALL jobs will be removed)",
					Value: -1,
				},
			},
		},
	}
)

func main() {
	app := cli.NewApp()

	app.Flags = AppFlags
	app.Commands = CliCommands

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}
}
