package node

import (
	"encoding/json"
	"path/filepath"

	"github.com/Spirals-Team/docker-g5k/libdockerg5k/swarm"
	"github.com/Spirals-Team/docker-g5k/libdockerg5k/weave"
	"github.com/Spirals-Team/docker-machine-driver-g5k/driver"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/log"
)

// NodeGlobalConfig contain needed information to provision a node
type NodeGlobalConfig struct {
	// libMachine client
	libMachineClient *libmachine.Client

	// g5k driver config (needed or Docker Machine will not work afterwards)
	g5kUsername          string
	g5kPassword          string
	g5kSite              string
	g5kImage             string
	g5kWalltime          string
	g5kSSHPrivateKeyPath string
	g5kSSHPublicKeyPath  string

	// Swarm configuration
	swarmStandaloneGlobalConfig *swarm.SwarmStandaloneGlobalConfig
	swarmModeGlobalConfig       *swarm.SwarmModeGlobalConfig

	// Weave networking
	weaveNetworkingEnabled bool
}

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
func (gc *NodeGlobalConfig) ProvisionNode(g5kJobID int, nodeName string, machineName string, isMaster bool) error {
	// create driver instance for libmachine
	driver := driver.NewDriver()

	// set g5k driver parameters
	driver.G5kUsername = gc.g5kUsername
	driver.G5kPassword = gc.g5kPassword
	driver.G5kSite = gc.g5kSite
	driver.G5kImage = gc.g5kImage
	driver.G5kWalltime = gc.g5kWalltime
	driver.G5kSSHPrivateKeyPath = gc.g5kSSHPrivateKeyPath
	driver.G5kSSHPublicKeyPath = gc.g5kSSHPublicKeyPath
	driver.G5kJobID = g5kJobID
	driver.G5kHostToProvision = nodeName

	// set base driver parameters
	driver.BaseDriver.MachineName = machineName
	driver.BaseDriver.StorePath = mcndirs.GetBaseDir()
	driver.BaseDriver.SSHKeyPath = driver.GetSSHKeyPath()

	// marshal configured driver
	data, err := json.Marshal(driver)
	if err != nil {
		return err
	}

	// create a new host config
	h, err := gc.libMachineClient.NewHost("g5k", data)
	if err != nil {
		return err
	}

	// mandatory, or driver will use bad paths for certificates
	h.HostOptions.AuthOptions = createHostAuthOptions(machineName)

	// set swarm options if Swarm standalone is enabled
	if gc.swarmStandaloneGlobalConfig != nil {
		h.HostOptions.SwarmOptions = gc.swarmStandaloneGlobalConfig.CreateSwarmStandaloneNodeConfig(nodeName, isMaster, true)
	}

	// provision the new machine
	if err := gc.libMachineClient.Create(h); err != nil {
		return err
	}

	// install and run Weave Net / Discovery if Weave networking mode and Swarm standalone are enabled
	if gc.weaveNetworkingEnabled && (gc.swarmStandaloneGlobalConfig != nil) {
		// run Weave Net
		log.Info("Running Weave Net...")
		if err := weave.RunWeaveNet(h); err != nil {
			return err
		}

		// run Weave Discovery
		log.Info("Running Weave Discovery...")
		if err := weave.RunWeaveDiscovery(h, gc.swarmStandaloneGlobalConfig.Discovery); err != nil {
			return err
		}
	}

	return nil
}
