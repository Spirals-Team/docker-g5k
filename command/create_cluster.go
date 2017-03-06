package command

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/swarm"

	"github.com/Spirals-Team/docker-machine-driver-g5k/driver"

	g5kswarm "github.com/Spirals-Team/docker-g5k/libdockerg5k/swarm"
	"github.com/Spirals-Team/docker-g5k/libdockerg5k/weave"
)

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
	runAgent := true
	// By default, exclude master node from Swarm pool, but can be overrided by swarm-master-join flag
	if isMaster && !c.cli.Bool("swarm-master-join") {
		runAgent = false
	}

	return &swarm.Options{
		IsSwarm:            true,
		Image:              c.cli.String("swarm-image"),
		Agent:              runAgent,
		Master:             isMaster,
		Discovery:          c.cli.String("swarm-discovery"),
		Address:            machineName,
		Host:               "tcp://0.0.0.0:3376",
		Strategy:           c.cli.String("swarm-strategy"),
		ArbitraryFlags:     c.cli.StringSlice("swarm-opt"),
		ArbitraryJoinFlags: c.cli.StringSlice("swarm-join-opt"),
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
	driver.G5kUsername = c.cli.String("g5k-username")
	driver.G5kPassword = c.cli.String("g5k-password")
	driver.G5kSite = c.cli.String("g5k-site")

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

	// install and run Weave Net / Discovery if Weave networking mode is enabled
	if c.cli.Bool("weave-networking") {
		// run Weave Net
		log.Info("Running Weave Net...")
		if err := weave.RunWeaveNet(h); err != nil {
			return err
		}

		// run Weave Discovery
		log.Info("Running Weave Discovery...")
		if err := weave.RunWeaveDiscovery(h, c.cli.String("swarm-discovery")); err != nil {
			return err
		}
	}

	return nil
}

// ProvisionNodes provision the nodes
func (c *Command) ProvisionNodes(nodes []string) error {
	// provision all deployed nodes
	var wg sync.WaitGroup
	for i, v := range nodes {
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

// checkCliParameters perform checks on CLI parameters
func (c *Command) checkCliParameters() error {
	// check username
	g5kUsername := c.cli.String("g5k-username")
	if g5kUsername == "" {
		return fmt.Errorf("You must provide your Grid5000 account username")
	}

	// check password
	g5kPassword := c.cli.String("g5k-password")
	if g5kPassword == "" {
		return fmt.Errorf("You must provide your Grid5000 account password")
	}

	// check site
	g5kSite := c.cli.String("g5k-site")
	if g5kSite == "" {
		return fmt.Errorf("You must provide a site to reserve the ressources on")
	}

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
		swarmDiscoveryToken, err := g5kswarm.GetNewSwarmDiscoveryToken()
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

	// TODO: Multi-sites reservation/deployment

	// reserve nodes via the Grid5000 API
	jobID, err := c.g5kAPI.ReserveNodes(c.cli.String("g5k-site"), c.cli.Int("g5k-nb-nodes"), c.cli.String("g5k-resource-properties"), c.cli.String("g5k-walltime"))
	if err != nil {
		return err
	}

	// submit new deployment
	deployedNodes, err := c.g5kAPI.DeployNodes(c.cli.String("g5k-site"), c.cli.String("g5k-ssh-public-key"), jobID, c.cli.String("g5k-image"))
	if err != nil {
		return err
	}

	// provision nodes
	if err := c.ProvisionNodes(deployedNodes); err != nil {
		return err
	}

	return nil
}
