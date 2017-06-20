package cluster

import (
	"encoding/json"
	"path/filepath"

	"github.com/Spirals-Team/docker-g5k/libdockerg5k/hostsmapping"
	"github.com/Spirals-Team/docker-g5k/libdockerg5k/weave"
	g5kdriver "github.com/Spirals-Team/docker-machine-driver-g5k/driver"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/auth"
)

// Node contain node specific informations
type Node struct {
	clusterConfig *GlobalConfig

	NodeName    string // Grid'5000 node hostname
	MachineName string // Docker Machine name

	// g5k driver
	G5kSite  string
	G5kJobID int

	// Docker Engine
	EngineOpt   []string
	EngineLabel []string
}

// createHostAuthOptions returns a configured AuthOptions for HostOptions struct
func (n *Node) createHostAuthOptions() *auth.Options {
	return &auth.Options{
		CertDir:          mcndirs.GetMachineCertDir(),
		CaCertPath:       filepath.Join(mcndirs.GetMachineCertDir(), "ca.pem"),
		CaPrivateKeyPath: filepath.Join(mcndirs.GetMachineCertDir(), "ca-key.pem"),
		ClientCertPath:   filepath.Join(mcndirs.GetMachineCertDir(), "cert.pem"),
		ClientKeyPath:    filepath.Join(mcndirs.GetMachineCertDir(), "key.pem"),
		ServerCertPath:   filepath.Join(mcndirs.GetMachineDir(), n.MachineName, "server.pem"),
		ServerKeyPath:    filepath.Join(mcndirs.GetMachineDir(), n.MachineName, "server-key.pem"),
		StorePath:        filepath.Join(mcndirs.GetMachineDir(), n.MachineName),
		ServerCertSANs:   nil,
	}
}

// Provision will install Docker Engine/Swarm and perform some configurations on the node
func (n *Node) Provision() error {
	// disable driver logs
	//log.SetErrWriter(ioutil.Discard)
	//log.SetOutWriter(ioutil.Discard)

	// create driver instance for libmachine
	driver := g5kdriver.NewDriver()

	// set g5k driver parameters
	driver.G5kUsername = n.clusterConfig.G5kUsername
	driver.G5kPassword = n.clusterConfig.G5kPassword
	driver.G5kSite = n.G5kSite
	driver.G5kImage = n.clusterConfig.G5kImage
	driver.G5kWalltime = n.clusterConfig.G5kWalltime
	driver.G5kJobID = n.G5kJobID
	driver.G5kHostToProvision = n.NodeName
	driver.SSHKeyPair = n.clusterConfig.SSHKeyPair

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
	h, err := n.clusterConfig.LibMachineClient.NewHost("g5k", data)
	if err != nil {
		return err
	}

	// set Docker Engine parameters
	h.HostOptions.EngineOptions.ArbitraryFlags = n.EngineOpt
	h.HostOptions.EngineOptions.Labels = n.EngineLabel

	// mandatory, or driver will use bad paths for certificates
	h.HostOptions.AuthOptions = n.createHostAuthOptions()

	// set swarm options if Swarm standalone is enabled
	if n.clusterConfig.SwarmStandaloneGlobalConfig != nil {
		h.HostOptions.SwarmOptions = n.clusterConfig.SwarmStandaloneGlobalConfig.CreateNodeConfig(n.NodeName, n.clusterConfig.SwarmMasterNode[n.MachineName], true)
	}

	// provision the new machine
	if err := n.clusterConfig.LibMachineClient.Create(h); err != nil {
		return err
	}

	// add all cluster nodes to the static lookup table of the host
	if err := hostsmapping.AddClusterHostsMapping(h, n.clusterConfig.HostsLookupTable); err != nil {
		return err
	}

	// install and run Weave Net / Discovery if Weave networking mode and Swarm standalone are enabled
	if n.clusterConfig.WeaveNetworkingEnabled && (n.clusterConfig.SwarmStandaloneGlobalConfig != nil) {
		// run Weave Net
		if err := weave.RunWeaveNet(h); err != nil {
			return err
		}

		// run Weave Discovery
		if err := weave.RunWeaveDiscovery(h, n.clusterConfig.SwarmStandaloneGlobalConfig.Discovery); err != nil {
			return err
		}
	}

	// Swarm mode
	if n.clusterConfig.SwarmModeGlobalConfig != nil {
		// check if cluster is already initialized
		if !n.clusterConfig.SwarmModeGlobalConfig.IsSwarmModeClusterInitialized() {
			// initialize Swarm mode cluster (only for bootstrap node)
			if err := n.clusterConfig.SwarmModeGlobalConfig.InitSwarmModeCluster(h); err != nil {
				return err
			}
		} else {
			// join the Swarm mode cluster
			if err := n.clusterConfig.SwarmModeGlobalConfig.JoinSwarmModeCluster(h, n.clusterConfig.SwarmMasterNode[n.MachineName]); err != nil {
				return err
			}
		}
	}

	return nil
}
