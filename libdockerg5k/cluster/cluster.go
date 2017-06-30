package cluster

import (
	"fmt"
	"sync"

	"net"

	"github.com/Spirals-Team/docker-g5k/libdockerg5k/swarm"
	"github.com/Spirals-Team/docker-g5k/libdockerg5k/zookeeper"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/ssh"
)

// GlobalConfig contains the cluster global configuration
type GlobalConfig struct {
	// Docker Machine
	LibMachineClient *libmachine.Client

	// Grid'5000 driver config (needed or some Docker Machine operations will not work afterwards)
	G5kUsername string
	G5kPassword string
	G5kImage    string
	G5kWalltime string
	SSHKeyPair  *ssh.KeyPair

	// Associates nodes IP address with Machine name
	HostsLookupTable map[string]string

	// Swarm configuration
	SwarmStandaloneGlobalConfig *swarm.SwarmStandaloneGlobalConfig
	SwarmModeGlobalConfig       *swarm.SwarmModeGlobalConfig
	SwarmMasterNode             []string

	// Weave networking
	WeaveNetworkingEnabled bool

	// Cluster storage
	UseZookeeperClusterStorage bool
}

// GenerateSSHKeyPair generate a new global SSH key
func (c *GlobalConfig) GenerateSSHKeyPair() error {
	sshKeyPair, err := ssh.NewKeyPair()
	if err != nil {
		return err
	}

	c.SSHKeyPair = sshKeyPair
	return nil
}

// Cluster represents the cluster
type Cluster struct {
	Config *GlobalConfig
	Nodes  map[string]*Node
}

// NewCluster create a new cluster using the given configuration
func NewCluster(config *GlobalConfig) *Cluster {
	return &Cluster{
		Config: config,
		Nodes:  make(map[string]*Node),
	}
}

// CreateNodes creates nodes from reservations
func (c *Cluster) CreateNodes(reservations map[string]int) {
	for site, count := range reservations {
		for i := 0; i < count; i++ {
			// generate machine name : {site}-{id}
			machineName := fmt.Sprintf("%s-%d", site, i)

			// store node configuration
			c.Nodes[machineName] = &Node{
				clusterConfig: c.Config,
				MachineName:   machineName,
				G5kSite:       site,
			}
		}
	}
}

// AllocateDeployedNodesToMachines allocate the deployed nodes to the Docker Machines
func (c *Cluster) AllocateDeployedNodesToMachines(site string, jobID int, deployedNodes []string) error {
	// create configuration for deployed nodes
	for i, n := range deployedNodes {
		// generate machine name : {site}-{id}
		machineName := fmt.Sprintf("%s-%d", site, i)

		// set driver parameters
		c.Nodes[machineName].NodeName = n
		c.Nodes[machineName].G5kJobID = jobID

		// lookup IP address of the node for static lookup table
		ip, err := net.LookupIP(n)
		if err != nil || len(ip) < 1 {
			return fmt.Errorf("Unable to lookup IP address for '%s' node: '%s'", n, err)
		}

		// set IP address of the machine in the static lookup table
		c.Config.HostsLookupTable[machineName] = ip[0].String()
	}

	return nil
}

// ProvisionNodes provision the nodes in the cluster (in parallel)
func (c *Cluster) ProvisionNodes() error {
	// if Swarm standalone is enabled, and no discovery method provided, deploy a Zookeeper instance for the cluster
	if (c.Config.SwarmStandaloneGlobalConfig != nil) && (c.Config.SwarmStandaloneGlobalConfig.Discovery == "") {
		log.Info("No Swarm cluster storage defined, Zookeeper will be deployed on each master nodes")

		// enable Zookeeper
		c.Config.UseZookeeperClusterStorage = true

		// set discovery string with zookeeper url
		c.Config.SwarmStandaloneGlobalConfig.Discovery = zookeeper.GenerateClusterStorageURL(c.Config.SwarmMasterNode, c.Config.HostsLookupTable)
	}

	// provision Swarm master/manager nodes (sequential)
	for _, k := range c.Config.SwarmMasterNode {
		log.Infof("Provisionning Swarm master/manager node '%s' ('%s')...", c.Nodes[k].NodeName, c.Nodes[k].MachineName)

		// error in Swarm master provisionning is fatal
		if err := c.Nodes[k].Provision(); err != nil {
			return fmt.Errorf("Error while provisionning Swarm master/manager node '%s' ('%s'): '%s'", c.Nodes[k].NodeName, c.Nodes[k].MachineName, err)
		}
	}

	log.Info("Provisionning nodes, it will take a few minutes...")

	// provision all deployed nodes (parallel)
	var wg sync.WaitGroup
	for _, n := range c.Nodes {
		// skip already provisionned Swarm master/manager
		if !n.isSwarmMaster() {
			wg.Add(1)
			go func(n *Node) {
				defer wg.Done()
				if err := n.Provision(); err != nil {
					log.Errorf("Error while provisionning node '%s' ('%s'): '%s'\n", n.NodeName, n.MachineName, err)
				}
			}(n)
		}
	}

	// wait nodes provisionning to finish
	wg.Wait()

	return nil
}
