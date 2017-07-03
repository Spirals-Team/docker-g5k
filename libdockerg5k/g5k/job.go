package g5k

import (
	"fmt"

	"github.com/Spirals-Team/docker-machine-driver-g5k/api"
)

// ReserveNodes allocate a new job with the required number of nodes on the given site, and returns the Job ID
func (g *G5K) ReserveNodes(site string, nbNodes int, resourceProperties string, walltime string) (int, error) {
	// create a new job request with given parameters
	jobReq := api.JobRequest{
		Resources:  fmt.Sprintf("nodes=%v,walltime=%s", nbNodes, walltime),
		Command:    "sleep 365d",
		Properties: resourceProperties,
		Types:      []string{"deploy"},
	}

	// get site API client
	siteAPI := g.getSiteAPI(site)

	// submit job request
	jobID, err := siteAPI.SubmitJob(jobReq)
	if err != nil {
		return 0, err
	}

	// wait until job reach 'ready' state
	if err := siteAPI.WaitUntilJobIsReady(jobID); err != nil {
		return 0, err
	}

	return jobID, nil
}
