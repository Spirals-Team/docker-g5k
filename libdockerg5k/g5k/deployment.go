package g5k

import (
	"github.com/Spirals-Team/docker-machine-driver-g5k/api"
)

// DeployNodes submit a deployment request and returns the deployed nodes hostname
func (g *G5K) DeployNodes(site string, sshPublicKey string, jobID int, image string) ([]string, error) {
	// get required site API client
	siteAPI := g.getSiteAPI(site)

	// get job informations
	job, err := siteAPI.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	// create a new deployment request
	deploymentReq := api.DeploymentRequest{
		Nodes:       job.Nodes,
		Environment: image,
		Key:         sshPublicKey,
	}

	// deploy environment
	deploymentID, err := siteAPI.SubmitDeployment(deploymentReq)
	if err != nil {
		return nil, err
	}

	// wait until deployment finish
	if err := siteAPI.WaitUntilDeploymentIsFinished(deploymentID); err != nil {
		return nil, err
	}

	// get deployment informations
	deployment, err := siteAPI.GetDeployment(deploymentID)
	if err != nil {
		return nil, err
	}

	return deployment.Nodes, nil
}
