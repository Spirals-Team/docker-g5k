package command

import (
	"encoding/json"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"

	"github.com/Spirals-Team/docker-machine-driver-g5k/driver"
)

var (
	// RemoveClusterCliCommand represent the CLI command "remove-cluster" with its flags
	RemoveClusterCliCommand = cli.Command{
		Name:   "remove-cluster",
		Usage:  "Remove a Docker Swarm cluster from the Grid'5000 infrastructure",
		Action: RunRemoveClusterCommand,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "g5k-job-id",
				Usage: "Only remove nodes related to the provided job ID (By default ALL nodes from ALL jobs will be removed)",
				Value: -1,
			},
		},
	}
)

// RemoveClusterCommand contain global parameters for the command "rm-cluster"
type RemoveClusterCommand struct {
	cli *cli.Context
}

// getG5kDriverConfig takes a raw driver and return a configured Driver structure
func (c *RemoveClusterCommand) getG5kDriverConfig(rawDriver []byte) (*driver.Driver, error) {
	// unmarshal driver configuration
	var drv driver.Driver
	if err := json.Unmarshal(rawDriver, &drv); err != nil {
		return nil, err
	}

	return &drv, nil
}

// RemoveCluster remove all nodes
func (c *RemoveClusterCommand) RemoveCluster() error {
	// create a new libmachine client
	client := libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir())
	defer client.Close()

	// load hosts from libmachine storage
	lst, _, err := persist.LoadAllHosts(client)
	if err != nil {
		return err
	}

	// store already deleted jobs to minimize API calls
	jobs := make(map[int]bool)

	// remove hosts from libmachine storage
	for _, h := range lst {
		// only remove Grid5000 nodes
		if h.DriverName == "g5k" {
			driverConfig, err := c.getG5kDriverConfig(h.RawDriver)
			if err != nil {
				log.Errorf("Cannot remove node '%s' : %s", h.Name, err)
			}

			// filter the nodes to remove on their job ID
			if jobID := c.cli.Int("g5k-job-id"); jobID != -1 {
				// skip nodes with different job ID than provided
				if driverConfig.G5kJobID != jobID {
					continue
				}
			}

			// check the job is already in the list of deleted jobs
			if _, exist := jobs[driverConfig.G5kJobID]; !exist {
				// send API call to kill job
				driverConfig.G5kAPI.KillJob(driverConfig.G5kJobID)

				// add job ID to list of deleted jobs
				jobs[driverConfig.G5kJobID] = true

				log.Infof("Job '%v' killed", driverConfig.G5kJobID)
			}

			// remove node from libmachine storage
			client.Remove(h.Name)

			log.Infof("Node '%s' removed", h.Name)
		}
	}

	return nil
}

// RunRemoveClusterCommand remove a cluster
func RunRemoveClusterCommand(cli *cli.Context) error {
	c := RemoveClusterCommand{cli: cli}
	return c.RemoveCluster()
}
