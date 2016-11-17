package command

import (
	"github.com/Spirals-Team/docker-machine-driver-g5k/api"
	"github.com/codegangsta/cli"
)

// Command struct contain common informations used between commands
type Command struct {
	cli *cli.Context
	api *api.Api

	g5kJobID        int
	g5kDeploymentID string
}

// NewCommandContext verify mandatory parameters and returns a new CommandContext
func NewCommandContext(cmd *cli.Context) (*Command, error) {
	return &Command{
		cli: cmd,
	}, nil
}

// RunCreateClusterCommand create a new cluster using parameters given in cli
func RunCreateClusterCommand(c *cli.Context) error {
	cmd, err := NewCommandContext(c)
	if err != nil {
		return err
	}

	if err := cmd.CreateCluster(); err != nil {
		return err
	}

	return nil
}

// RunRemoveClusterCommand remove an existing cluster using parameters given in cli
func RunRemoveClusterCommand(c *cli.Context) error {
	cmd, err := NewCommandContext(c)
	if err != nil {
		return err
	}

	if err := cmd.RemoveCluster(); err != nil {
		return err
	}

	return nil
}
