package command

import (
	"strconv"

	"github.com/Spirals-Team/docker-machine-driver-g5k/api"
)

// Command struct contain all common informations used between commands
type Command struct {
	G5kAPI *api.Api

	G5kJobID        int
	G5kDeploymentID string

	G5kUsername           string
	G5kPassword           string
	G5kSite               string
	G5kWalltime           string
	G5kSSHPrivateKeyPath  string
	G5kSSHPublicKeyPath   string
	G5kImage              string
	G5kResourceProperties string
	G5kNbNodes            int

	SwarmDiscoveryToken string
}

// ParseArguments will parse command line arguments an set data in common struct (later)
func (c *Command) ParseArguments(args []string) error {
	// TODO: really parse arguments from cli
	c.G5kImage = "jessie-x64-min"
	c.G5kUsername = args[0]
	c.G5kPassword = args[1]
	c.G5kSite = args[2]
	c.G5kWalltime = args[3]
	c.G5kSSHPrivateKeyPath = args[4]
	c.G5kSSHPublicKeyPath = args[4] + ".pub"

	c.G5kAPI = api.NewApi(c.G5kUsername, c.G5kPassword, c.G5kSite, c.G5kImage)

	c.SwarmDiscoveryToken = args[5]

	var err error
	if c.G5kNbNodes, err = strconv.Atoi(args[6]); err != nil {
		return err
	}

	return nil
}
