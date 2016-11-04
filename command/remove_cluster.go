package command

import (
	"encoding/json"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"

	"github.com/Spirals-Team/docker-machine-driver-g5k/driver"
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

	// store already deleted jobs to minimize API calls
	jobs := make(map[int]bool)

	// remove hosts from libmachine storage
	for _, h := range lst {
		// only remove Grid5000 nodes
		if h.DriverName == "g5k" {
			driverConfig, err := getG5kDriverConfig(h.RawDriver)
			if err != nil {
				log.Errorf("Cannot remove node '%s' : %s", h.Name, err)
			}

			// check the job is already in the list of deleted jobs
			if _, exist := jobs[driverConfig.G5kJobID]; !exist {
				// send API call to kill job
				c.api.KillJob(driverConfig.G5kJobID)

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

// getG5kDriverConfig takes a raw driver and return a configured Driver structure
func getG5kDriverConfig(rawDriver []byte) (*driver.Driver, error) {
	// unmarshal driver configuration
	var drv driver.Driver
	if err := json.Unmarshal(rawDriver, &drv); err != nil {
		return nil, err
	}

	return &drv, nil
}
