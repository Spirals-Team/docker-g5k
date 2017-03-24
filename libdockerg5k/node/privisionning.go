package node

import (
	"encoding/json"
	"path/filepath"

	"github.com/Spirals-Team/docker-g5k/libdockerg5k/weave"
	g5kdriver "github.com/Spirals-Team/docker-machine-driver-g5k/driver"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/log"
)

// createHostAuthOptions returns a configured AuthOptions for HostOptions struct
func createHostAuthOptions(machineName string) *auth.Options {
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

// ProvisionNode will install Docker and configure Swarm (if enabled) on the node
func (n *Node) ProvisionNode() error {
	// create driver instance for libmachine
	driver := g5kdriver.NewDriver()

	// set g5k driver parameters
	driver.G5kUsername = n.GlobalConfig.G5kUsername
	driver.G5kPassword = n.GlobalConfig.G5kPassword
	driver.G5kSite = n.G5kSite
	driver.G5kImage = n.GlobalConfig.G5kImage
	driver.G5kWalltime = n.GlobalConfig.G5kWalltime
	driver.G5kJobID = n.G5kJobID
	driver.G5kHostToProvision = n.NodeName
	driver.SSHKeyPair = n.GlobalConfig.SSHKeyPair

	// set base driver parameters
	driver.BaseDriver.MachineName = n.MachineName
	driver.BaseDriver.StorePath = mcndirs.GetBaseDir()
	driver.BaseDriver.SSHKeyPath = driver.GetSSHKeyPath()

	// marshal configured driver
	data, err := json.Marshal(driver)
	if err != nil {
		return err
	}

	// create a new host config
	h, err := n.GlobalConfig.LibMachineClient.NewHost("g5k", data)
	if err != nil {
		return err
	}

	// mandatory, or driver will use bad paths for certificates
	h.HostOptions.AuthOptions = createHostAuthOptions(n.MachineName)

	// set swarm options if Swarm standalone is enabled
	if n.GlobalConfig.SwarmStandaloneGlobalConfig != nil {
		h.HostOptions.SwarmOptions = n.GlobalConfig.SwarmStandaloneGlobalConfig.CreateNodeConfig(n.NodeName, n.GlobalConfig.SwarmMasterNode[n.MachineName], true)
	}

	// provision the new machine
	if err := n.GlobalConfig.LibMachineClient.Create(h); err != nil {
		return err
	}

	// install and run Weave Net / Discovery if Weave networking mode and Swarm standalone are enabled
	if n.GlobalConfig.WeaveNetworkingEnabled && (n.GlobalConfig.SwarmStandaloneGlobalConfig != nil) {
		// run Weave Net
		log.Info("Running Weave Net...")
		if err := weave.RunWeaveNet(h); err != nil {
			return err
		}

		// run Weave Discovery
		log.Info("Running Weave Discovery...")
		if err := weave.RunWeaveDiscovery(h, n.GlobalConfig.SwarmStandaloneGlobalConfig.Discovery); err != nil {
			return err
		}
	}

	// Swarm mode
	if n.GlobalConfig.SwarmModeGlobalConfig != nil {
		// check if cluster is already initialized
		if !n.GlobalConfig.SwarmModeGlobalConfig.IsSwarmModeClusterInitialized() {
			// initialize Swarm mode cluster (only for bootstrap node)
			if err := n.GlobalConfig.SwarmModeGlobalConfig.InitSwarmModeCluster(h); err != nil {
				return err
			}
		} else {
			// join the Swarm mode cluster
			if err := n.GlobalConfig.SwarmModeGlobalConfig.JoinSwarmModeCluster(h, n.GlobalConfig.SwarmMasterNode[n.MachineName]); err != nil {
				return err
			}
		}
	}

	return nil
}
