package g5k

import (
	"fmt"

	"github.com/Spirals-Team/docker-machine-driver-g5k/api"
)

// ReserveNodes allocate a new job with the required number of nodes on the given site, and returns the Job ID
func (g *G5K) ReserveNodes(site string, nbNodes int, resourceProperties string, walltime string) (int, error) {
	// convert walltime to seconds
	seconds, err := api.ConvertDuration(walltime)
	if err != nil {
		return -1, err
	}

	// create a new job request with given parameters
	jobReq := api.JobRequest{
		Resources:  fmt.Sprintf("nodes=%v,walltime=%s", nbNodes, walltime),
		Command:    fmt.Sprintf("sleep %v", seconds),
		Properties: resourceProperties,
		Types:      []string{"deploy"},
	}

	// get site API client
	siteAPI := g.getSiteAPI(site)

	// submit job request
	g5kJobID, err := siteAPI.SubmitJob(jobReq)
	if err != nil {
		return -1, err
	}

	return g5kJobID, nil
}

// WaitUntilJobIsReady wait until job reach 'ready' state
func (g *G5K) WaitUntilJobIsReady(site string, jobID int) error {
	siteAPI := g.getSiteAPI(site)

	if err := siteAPI.WaitUntilJobIsReady(jobID); err != nil {
		return err
	}

	return nil
}
