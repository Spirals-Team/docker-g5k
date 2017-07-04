package command

import (
	"strconv"

	"fmt"

	"strings"

	"github.com/Songmu/prompter"
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
		Usage:   "Remove a Docker cluster from the Grid'5000 infrastructure",
		Action:  RunRemoveClusterCommand,
		Flags: []cli.Flag{
			cli.BoolFlag{
				EnvVar: "G5K_RM_NO_CONFIRM",
				Name:   "no-confirm",
				Usage:  "Disable confirmation before removing machines",
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
	if c.cli.NArg() < 1 {
		return fmt.Errorf("You must provide the job ID of the nodes you want to remove")
	}

	return nil
}

// RemoveCluster remove all nodes
func (c *RemoveClusterCommand) RemoveCluster() error {
	// create a new libmachine client
	client := libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir())
	defer client.Close()

	// store jobs ID to kill
	jobsToKill := make(map[int]bool)

	for _, arg := range c.cli.Args() {
		// convert argument to int
		jobID, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("The given parameter '%s' is not a valid job ID", arg)
		}

		// append job ID to list of jobs to kill
		jobsToKill[jobID] = true
	}

	// if confirmation is enabled (default behavior)
	if !c.cli.Bool("no-confirm") {
		// warn user before starting
		log.Infof("About to remove all associated machine(s) for job(s) ID : %s", strings.Join(c.cli.Args(), ", "))
		log.Warn("WARNING: This action terminate the resource reservation(s) and the node(s) will be unavailable !")

		// ask for confirmation
		if !prompter.YN("Are you sure?", false) {
			return fmt.Errorf("The operation was canceled by the user")
		}
	}

	// load hosts from libmachine storage
	lst, _, err := persist.LoadAllHosts(client)
	if err != nil {
		return err
	}

	// store already deleted jobs to minimize API calls
	killedJobs := make(map[int]bool)

	// remove hosts from libmachine storage
	for _, h := range lst {
		// only remove Grid5000 nodes
		if h.DriverName == "g5k" {
			// get machine's driver configuration
			driverConfig, err := GetG5kDriverConfig(h.RawDriver)
			if err != nil {
				log.Errorf("Cannot remove node '%s' : %s", h.Name, err)
			}

			// skip nodes with Job ID not in the list of jobs to kill
			if _, exist := jobsToKill[driverConfig.G5kJobID]; !exist {
				continue
			}

			// check the job is already in the list of deleted jobs
			if _, exist := killedJobs[driverConfig.G5kJobID]; !exist {
				// send API call to kill job
				driverConfig.G5kAPI.KillJob(driverConfig.G5kJobID)

				// add job ID to list of deleted jobs
				killedJobs[driverConfig.G5kJobID] = true

				log.Infof("Job '%v' killed", driverConfig.G5kJobID)
			}

			// remove node from libmachine storage
			if err := client.Remove(h.Name); err != nil {
				return fmt.Errorf("Operation aborted. Error while removing '%s' machine: '%s'", h.Name, err)
			}

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
