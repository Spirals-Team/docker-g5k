package g5k

import (
	"io/ioutil"

	"github.com/Spirals-Team/docker-machine-driver-g5k/api"
)

// DeployNodes submit a deployment request and returns the Deployment ID
func (g *G5K) DeployNodes(site string, sshPublicKeyPath string, jobID int, image string) (string, error) {
	// reading ssh public key file
	pubkey, err := ioutil.ReadFile(sshPublicKeyPath)
	if err != nil {
		return "", err
	}

	// get required site API client
	siteAPI := g.getSiteAPI(site)

	// get job informations
	job, err := siteAPI.GetJob(jobID)
	if err != nil {
		return "", err
	}

	// creating a new deployment request
	deploymentReq := api.DeploymentRequest{
		Nodes:       job.Nodes,
		Environment: image,
		Key:         string(pubkey),
	}

	// deploy environment
	g5kDeploymentID, err := siteAPI.SubmitDeployment(deploymentReq)
	if err != nil {
		return "", err
	}

	return g5kDeploymentID, nil
}

// WaitUntilDeploymentIsFinished wait until deployment finish
func (g *G5K) WaitUntilDeploymentIsFinished(site string, deploymentID string) error {
	siteAPI := g.getSiteAPI(site)

	if err := siteAPI.WaitUntilDeploymentIsFinished(deploymentID); err != nil {
		return err
	}

	return nil
}
