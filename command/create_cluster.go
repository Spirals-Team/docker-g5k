package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/swarm"

	"github.com/Spirals-Team/docker-machine-driver-g5k/api"
	"github.com/Spirals-Team/docker-machine-driver-g5k/driver"

	"github.com/docker/swarm/discovery/token"
)

// AllocateNodes allocate a new job with multiple nodes
func (c *Command) AllocateNodes() error {
	// convert walltime to seconds
	seconds, err := api.ConvertDuration(c.cli.String("g5k-walltime"))
	if err != nil {
		return err
	}

	// create a new job request
	jobReq := api.JobRequest{
		Resources:  fmt.Sprintf("nodes=%v,walltime=%s", c.cli.Int("g5k-nb-nodes"), c.cli.String("g5k-walltime")),
		Command:    fmt.Sprintf("sleep %v", seconds),
		Properties: c.cli.String("g5k-resource-properties"),
		Types:      []string{"deploy"},
	}

	// submit job request
	c.g5kJobID, err = c.api.SubmitJob(jobReq)
	if err != nil {
		return err
	}

	return nil
}

// DeployNodes submit a deployment request
func (c *Command) DeployNodes() error {
	// reading ssh public key file
	pubkey, err := ioutil.ReadFile(c.cli.String("g5k-ssh-public-key"))
	if err != nil {
		return err
	}

	// get job informations
	job, err := c.api.GetJob(c.g5kJobID)
	if err != nil {
		return err
	}

	// creating a new deployment request
	deploymentReq := api.DeploymentRequest{
		Nodes:       job.Nodes,
		Environment: c.cli.String("g5k-image"),
		Key:         string(pubkey),
	}

	// deploy environment
	c.g5kDeploymentID, err = c.api.SubmitDeployment(deploymentReq)
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
	joinFlags := c.cli.StringSlice("swarm-join-opt")

	// Advertise to the Weave Proxy port when a node join a Swarm cluster if Weave networking is enabled
	if c.cli.Bool("weave-networking") {
		joinFlags = append(joinFlags, fmt.Sprintf("advertise=%s:12375", machineName))
	}

	return &swarm.Options{
		IsSwarm:            true,
		Image:              c.cli.String("swarm-image"),
		Agent:              !isMaster, // exclude Swarm master from Swarm Pool
		Master:             isMaster,
		Discovery:          c.cli.String("swarm-discovery"),
		Address:            machineName,
		Host:               "tcp://0.0.0.0:3376",
		Strategy:           c.cli.String("swarm-strategy"),
		ArbitraryFlags:     c.cli.StringSlice("swarm-opt"),
		ArbitraryJoinFlags: joinFlags,
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
	driver.G5kUsername = c.cli.GlobalString("g5k-username")
	driver.G5kPassword = c.cli.GlobalString("g5k-password")
	driver.G5kSite = c.cli.GlobalString("g5k-site")

	driver.G5kImage = c.cli.String("g5k-image")
	driver.G5kWalltime = c.cli.String("g5k-walltime")
	driver.G5kSSHPrivateKeyPath = c.cli.String("g5k-ssh-private-key")
	driver.G5kSSHPublicKeyPath = c.cli.String("g5k-ssh-public-key")

	driver.G5kHostToProvision = nodeName
	driver.G5kJobID = c.g5kJobID

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
	deployment, err := c.api.GetDeployment(c.g5kDeploymentID)
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

// generateNewSwarmDiscoveryToken get a new Docker Swarm discovery token from Docker Hub
func generateNewSwarmDiscoveryToken() (string, error) {
	// init Discovery structure
	discovery := token.Discovery{}
	discovery.Initialize("token", 0, 0, nil)

	// get a new discovery token from Docker Hub
	swarmToken, err := discovery.CreateCluster()
	if err != nil {
		return "", err
	}

	return swarmToken, nil
}

// checkSshKeyFiles check if
func (c *Command) checkCliParameters() error {
	// check ssh private key
	sshPrivKey := c.cli.String("g5k-ssh-private-key")
	if sshPrivKey == "" {
		return fmt.Errorf("You must provide your SSH private key path")
	}

	// check if private key file exist
	if _, err := os.Stat(sshPrivKey); os.IsNotExist(err) {
		return fmt.Errorf("Your ssh private key file does not exist in : '%s'", sshPrivKey)
	}

	// check ssh public key, set it to '<privKey>.pub' if not set
	sshPubKey := c.cli.String("g5k-ssh-public-key")
	if sshPubKey == "" {
		c.cli.Set("g5k-ssh-public-key", fmt.Sprintf("%s.pub", sshPrivKey))
		sshPubKey = c.cli.String("g5k-ssh-public-key")
	}

	// check if public key file exist
	if _, err := os.Stat(sshPubKey); os.IsNotExist(err) {
		return fmt.Errorf("Your ssh public key file does not exist in : '%s'", sshPubKey)
	}

	// check Docker Swarm discovery
	swarmDiscovery := c.cli.String("swarm-discovery")
	if swarmDiscovery == "" {
		swarmDiscoveryToken, err := generateNewSwarmDiscoveryToken()
		if err != nil {
			return err
		}

		// set discovery token in CLI context
		c.cli.Set("swarm-discovery", fmt.Sprintf("token://%s", swarmDiscoveryToken))

		log.Infof("New Swarm discovery token generated : '%s'", swarmDiscoveryToken)
	}

	// check Docker Swarm image
	swarmImage := c.cli.String("swarm-image")
	if swarmImage == "" {
		return fmt.Errorf("You must provide a Swarm image")
	}

	// check Docker Swarm strategy
	swarmStrategy := c.cli.String("swarm-strategy")
	if swarmStrategy == "" {
		return fmt.Errorf("You must provide a Swarm strategy")
	}

	return nil
}

// CreateCluster create nodes in docker-machine
func (c *Command) CreateCluster() error {
	// create a new libmachine client
	client := libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir())
	defer client.Close()

	// check cli parameters
	if err := c.checkCliParameters(); err != nil {
		return err
	}

	// submit new job
	if err := c.AllocateNodes(); err != nil {
		return err
	}

	// wait until job is running
	c.api.WaitUntilJobIsReady(c.g5kJobID)

	// submit new deployment
	if err := c.DeployNodes(); err != nil {
		return err
	}

	// wait until deployment is finished
	c.api.WaitUntilDeploymentIsFinished(c.g5kDeploymentID)

	// provision nodes
	if err := c.ProvisionNodes(); err != nil {
		return err
	}

	return nil
}
