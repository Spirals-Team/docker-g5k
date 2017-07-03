package command

import (
	"sort"

	"fmt"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"
)

var (
	// RemoveClusterCliCommand represent the CLI command "remove-cluster" with its flags
	RemoveClusterCliCommand = cli.Command{
		Name:    "remove-cluster",
		Aliases: []string{"rm-cluster", "rm", "r"},
		Usage:   "Remove a Docker Swarm cluster from the Grid'5000 infrastructure",
		Action:  RunRemoveClusterCommand,
		Flags: []cli.Flag{
			cli.IntSliceFlag{
				EnvVar: "G5K_JOB_ID",
				Name:   "g5k-job-id",
				Usage:  "Only remove nodes related to the provided job ID",
			},
		},
	}
)

// RemoveClusterCommand contain global parameters for the command "rm-cluster"
type RemoveClusterCommand struct {
	cli *cli.Context
}

func (c *RemoveClusterCommand) checkCliParameters() error {
	// check job ID
	if len(c.cli.IntSlice("g5k-job-id")) < 1 {
		return fmt.Errorf("You must provide the job ID of the nodes you want to remove")
	}

	return nil
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

	// sort job ID to kill
	sort.Ints(c.cli.IntSlice("g5k-job-id"))

	// store already deleted jobs to minimize API calls
	deletedJobs := make(map[int]bool)

	// remove hosts from libmachine storage
	for _, h := range lst {
		// only remove Grid5000 nodes
		if h.DriverName == "g5k" {
			driverConfig, err := GetG5kDriverConfig(h.RawDriver)
			if err != nil {
				log.Errorf("Cannot remove node '%s' : %s", h.Name, err)
			}

			// skip nodes with Job ID not in the list of jobs to kill
			if sort.SearchInts(c.cli.IntSlice("g5k-job-id"), driverConfig.G5kJobID) != len(c.cli.IntSlice("g5k-job-id")) {
				continue
			}

			// check the job is already in the list of deleted jobs
			if _, exist := deletedJobs[driverConfig.G5kJobID]; !exist {
				// send API call to kill job
				driverConfig.G5kAPI.KillJob(driverConfig.G5kJobID)

				// add job ID to list of deleted jobs
				deletedJobs[driverConfig.G5kJobID] = true

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

	// check CLI parameters
	if err := c.checkCliParameters(); err != nil {
		return err
	}

	return c.RemoveCluster()
}
