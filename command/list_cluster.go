package command

import (
	"fmt"
	"os"
	"text/tabwriter"

	"strings"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/persist"
)

var (
	// ListClusterCliCommand represent the CLI command "list-cluster" with its flags
	ListClusterCliCommand = cli.Command{
		Name:    "list-cluster",
		Aliases: []string{"ls-cluster", "ls", "l"},
		Usage:   "List all clusters and their number of nodes",
		Action:  RunListClusterCommand,
	}
)

// ListClusterCommand contain global parameters for the command "ls-cluster"
type ListClusterCommand struct {
	cli *cli.Context
}

// ListCluster list all clusters
func (c *ListClusterCommand) ListCluster() error {
	// create a new libmachine client
	client := libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir())
	defer client.Close()

	// load hosts from libmachine storage
	lst, _, err := persist.LoadAllHosts(client)
	if err != nil {
		return err
	}

	// store machines count per jobs
	machinesPerJobs := make(map[int][]string)

	// count machines per Job
	for _, machine := range lst {
		// only catch Grid'5000 nodes
		if machine.DriverName == "g5k" {
			// get machine driver configuration
			driverConfig, err := GetG5kDriverConfig(machine.RawDriver)
			if err != nil {
				continue
			}

			// add machine to the list of machines for its job ID
			machinesPerJobs[driverConfig.G5kJobID] = append(machinesPerJobs[driverConfig.G5kJobID], machine.Name)
		}
	}

	// output writer with automatic tab handling
	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)

	// print header
	fmt.Fprintf(w, "JOB ID\tNUMBER OF MACHINE(S)\tMACHINE(S) NAME\n")

	// print jobs informations (id, number of machines, machines name)
	for j, n := range machinesPerJobs {
		fmt.Fprintf(w, "%d\t%d\t%s\n", j, len(n), strings.Join(n, ", "))
	}

	// flush output buffer
	w.Flush()

	return nil
}

// RunListClusterCommand list all clusters you reserved
func RunListClusterCommand(cli *cli.Context) error {
	c := ListClusterCommand{cli: cli}
	return c.ListCluster()
}
