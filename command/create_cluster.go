package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/swarm"

	"github.com/Spirals-Team/docker-machine-driver-g5k/api"
	"github.com/Spirals-Team/docker-machine-driver-g5k/driver"
)

// AllocateNodes allocate a new job with multiple nodes
func (c *Command) AllocateNodes() error {
	// convert walltime to seconds
	seconds, err := api.ConvertDuration(c.G5kWalltime)
	if err != nil {
		return err
	}

	// create a new job request
	jobReq := api.JobRequest{
		Resources:  fmt.Sprintf("nodes=%v,walltime=%s", c.G5kNbNodes, c.G5kWalltime),
		Command:    fmt.Sprintf("sleep %v", seconds),
		Properties: c.G5kResourceProperties,
		Types:      []string{"deploy"},
	}

	// submit job request
	c.G5kJobID, err = c.G5kAPI.SubmitJob(jobReq)
	if err != nil {
		return err
	}

	return nil
}

// DeployNodes deploy the nodes in a job
func (c *Command) DeployNodes() error {
	// reading ssh public key file
	pubkey, err := ioutil.ReadFile(c.G5kSSHPublicKeyPath)
	if err != nil {
		return err
	}

	// get job informations
	job, err := c.G5kAPI.GetJob(c.G5kJobID)
	if err != nil {
		return err
	}

	// creating a new deployment request
	deploymentReq := api.DeploymentRequest{
		Nodes:       job.Nodes,
		Environment: c.G5kImage,
		Key:         string(pubkey),
	}

	// deploy environment
	c.G5kDeploymentID, err = c.G5kAPI.SubmitDeployment(deploymentReq)
	if err != nil {
		return err
	}

	return nil
}

// createHostAuthOptions returns a configured AuthOptions for HostOptions struct
func (c *Command) createHostAuthOptions(machineName string) *auth.Options {
	return &auth.Options{
		CertDir:          mcndirs.GetMachineCertDir(),
		CaCertPath:       filepath.Join(mcndirs.GetMachineCertDir(), "ca.pem"),
		CaPrivateKeyPath: filepath.Join(mcndirs.GetMachineCertDir(), "ca-key.pem"),
		ClientCertPath:   filepath.Join(mcndirs.GetMachineCertDir(), "cert.pem"),
		ClientKeyPath:    filepath.Join(mcndirs.GetMachineCertDir(), "key.pem"),
		ServerCertPath:   filepath.Join(mcndirs.GetMachineDir(), machineName, "server.pem"),
		ServerKeyPath:    filepath.Join(mcndirs.GetMachineDir(), machineName, "server-key.pem"),
		StorePath:        filepath.Join(mcndirs.GetMachineDir(), machineName),
		ServerCertSANs:   nil,
	}
}

// createHostSwarmOptions returns a configured SwarmOptions for HostOptions struct
func (c *Command) createHostSwarmOptions(machineName string, isMaster bool) *swarm.Options {
	return &swarm.Options{
		IsSwarm: true,
		Image:   "swarm:latest",
		// Agent:          !isMaster, to exclude Swarm master from Swarm Pool
		Agent:          true,
		Master:         isMaster,
		Discovery:      "token://" + c.SwarmDiscoveryToken,
		Address:        machineName,
		Host:           "tcp://0.0.0.0:3376",
		Strategy:       "spread",
		ArbitraryFlags: nil,
		// Weave: ArbitraryJoinFlags: []string{"advertise=0.0.0.0:12375"},
		ArbitraryJoinFlags: nil,
		IsExperimental:     false,
	}
}

func (c *Command) provisionNode(nodeName string, isSwarmMaster bool) error {
	// create a new libmachine client
	client := libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir())
	defer client.Close()

	// create driver instance for libmachine
	driver := driver.NewDriver()

	// set g5k driver parameters
	driver.G5kImage = c.G5kImage
	driver.G5kUsername = c.G5kUsername
	driver.G5kPassword = c.G5kPassword
	driver.G5kSite = c.G5kSite
	driver.G5kWalltime = c.G5kWalltime
	driver.G5kSSHPrivateKeyPath = c.G5kSSHPrivateKeyPath
	driver.G5kSSHPublicKeyPath = c.G5kSSHPublicKeyPath
	driver.G5kHostToProvision = nodeName
	driver.G5kJobID = c.G5kJobID

	// set base driver parameters
	driver.BaseDriver.MachineName = nodeName
	driver.BaseDriver.StorePath = mcndirs.GetBaseDir()
	driver.BaseDriver.SSHKeyPath = driver.GetSSHKeyPath()

	// marshal configured driver
	data, err := json.Marshal(driver)
	if err != nil {
		return err
	}

	// create a new host config
	h, err := client.NewHost("g5k", data)
	if err != nil {
		return err
	}

	// mandatory, or driver will use bad paths
	h.HostOptions.AuthOptions = c.createHostAuthOptions(nodeName)

	// set swarm options
	h.HostOptions.SwarmOptions = c.createHostSwarmOptions(nodeName, isSwarmMaster)

	// provision the new machine
	if err := client.Create(h); err != nil {
		return err
	}

	return nil
}

// ProvisionNodes provision the nodes
func (c *Command) ProvisionNodes() error {
	// get deployment informations
	deployment, err := c.G5kAPI.GetDeployment(c.G5kDeploymentID)
	if err != nil {
		return err
	}

	// provision all deployed nodes
	var wg sync.WaitGroup
	for i, v := range deployment.Nodes {
		wg.Add(1)
		go func(nodeID int, nodeName string) {
			defer wg.Done()

			// first node will be the swarm master
			if nodeID == 0 {
				c.provisionNode(nodeName, true)
			} else {
				c.provisionNode(nodeName, false)
			}

		}(i, v)
	}

	// wait nodes provisionning to finish
	wg.Wait()

	return nil
}

// CreateCluster create nodes in docker-machine
func (c *Command) CreateCluster() error {
	// create a new libmachine client
	client := libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir())
	defer client.Close()

	// submit new job
	if err := c.AllocateNodes(); err != nil {
		return err
	}

	// wait until job is running
	c.G5kAPI.WaitUntilJobIsReady(c.G5kJobID)

	// submit new deployment
	if err := c.DeployNodes(); err != nil {
		return err
	}

	// wait until deployment is finished
	c.G5kAPI.WaitUntilDeploymentIsFinished(c.G5kDeploymentID)

	// provision nodes
	if err := c.ProvisionNodes(); err != nil {
		return err
	}

	return nil
}

// printHostConfig print the json config of a Host (used for debug)
func printHostConfig(h *host.Host) {
	host, err := json.Marshal(h)
	if err != nil {
		log.Error(err)
		return
	}

	fmt.Println(string(host))
}
