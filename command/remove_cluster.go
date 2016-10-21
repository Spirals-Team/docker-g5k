package command

import (
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/persist"
)

// RemoveCluster remove all nodes
func (c *Command) RemoveCluster() error {
	// create a new libmachine client
	client := libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir())
	defer client.Close()

	// load hosts from libmachine storage
	lst, _, err := persist.LoadAllHosts(client)
	if err != nil {
		return err
	}

	// TODO: Kill Grid5000 job

	// remove hosts from libmachine storage
	for _, v := range lst {
		client.Remove(v.Name)
	}

	return nil
}
