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
	AppFlags = []cli.Flag{
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
			Name:  "g5k-site",
			Usage: "Site to reserve the resources on",
			Value: "",
		},
	}

	// CliCommands stores cli commands and their flags
	CliCommands = []cli.Command{
		{
			Name:   "create-cluster",
			Usage:  "Create a new Docker Swarm cluster on Grid5000",
			Action: command.RunCreateClusterCommand,
			Flags: []cli.Flag{
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
					Usage: "Path of your ssh public key",
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

				cli.IntFlag{
					Name:  "g5k-nb-nodes",
					Usage: "Number of nodes to allocate",
					Value: 3,
				},

				cli.StringFlag{
					Name:  "swarm-discovery-token",
					Usage: "Discovery token to use for joining a Swarm cluster",
					Value: "",
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
