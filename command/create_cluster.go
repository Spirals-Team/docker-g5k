package command

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"

	"github.com/Spirals-Team/docker-g5k/libdockerg5k/g5k"
	"github.com/Spirals-Team/docker-g5k/libdockerg5k/node"
	"github.com/Spirals-Team/docker-g5k/libdockerg5k/swarm"
)

var (
	// CreateClusterCliCommand represent the CLI command "create-cluster" with its flags
	CreateClusterCliCommand = cli.Command{
		Name:   "create-cluster",
		Usage:  "Create a new Docker Swarm cluster on the Grid'5000 infrastructure",
		Action: RunCreateClusterCommand,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "g5k-username",
				Usage: "Your Grid5000 account username",
				Value: "",
			},

			cli.StringFlag{
				Name:  "g5k-password",
				Usage: "Your Grid5000 account password",
				Value: "",
			},

			cli.StringSliceFlag{
				Name:  "g5k-reserve-nodes",
				Usage: "Reserve nodes on a site (ex: lille:24)",
			},

			cli.StringFlag{
				Name:  "g5k-walltime",
				Usage: "Machine's lifetime (HH:MM:SS)",
				Value: "1:00:00",
			},

			cli.StringFlag{
				Name:  "g5k-image",
				Usage: "Name of the image to deploy",
				Value: "jessie-x64-min",
			},

			cli.StringFlag{
				Name:  "g5k-resource-properties",
				Usage: "Resource selection with OAR properties (SQL format)",
				Value: "",
			},

			cli.StringSliceFlag{
				Name:  "engine-opt",
				Usage: "Specify arbitrary flags to include on the selected node(s) engine (site-id:flag=value)",
			},

			cli.StringSliceFlag{
				Name:  "engine-label",
				Usage: "Specify labels for the selected node(s) engine (site-id:labelname=labelvalue)",
			},

			cli.StringSliceFlag{
				Name:  "swarm-master",
				Usage: "Select node(s) to be promoted to Swarm Master(standalone)/Manager(Mode)",
			},

			cli.BoolFlag{
				Name:  "swarm-mode-enable",
				Usage: "Create a Swarm mode cluster",
			},

			cli.BoolFlag{
				Name:  "swarm-standalone-enable",
				Usage: "Create a Swarm standalone cluster",
			},

			cli.StringFlag{
				Name:  "swarm-standalone-discovery",
				Usage: "Discovery service to use with Swarm",
				Value: "",
			},

			cli.StringFlag{
				Name:  "swarm-standalone-image",
				Usage: "Specify Docker image to use for Swarm",
				Value: "swarm:latest",
			},

			cli.StringFlag{
				Name:  "swarm-standalone-strategy",
				Usage: "Define a default scheduling strategy for Swarm",
				Value: "spread",
			},

			cli.StringSliceFlag{
				Name:  "swarm-standalone-opt",
				Usage: "Define arbitrary flags for Swarm master (can be provided multiple times)",
			},

			cli.StringSliceFlag{
				Name:  "swarm-standalone-join-opt",
				Usage: "Define arbitrary flags for Swarm join (can be provided multiple times)",
			},

			cli.BoolFlag{
				Name:  "weave-networking",
				Usage: "Use Weave for networking (Only if Swarm standalone is enabled)",
			},
		},
	}
)

// CreateClusterCommand contain global parameters for the command "create-cluster"
type CreateClusterCommand struct {
	cli               *cli.Context
	nodesReservation  map[string]int
	swarmMasterNodes  map[string]bool
	nodesGlobalConfig *node.GlobalConfig
}

// parseReserveNodesFlag parse the nodes reservation flag (site):(number of nodes)
func (c *CreateClusterCommand) parseReserveNodesFlag() error {
	// initialize nodes reservation map
	c.nodesReservation = make(map[string]int)

	// TODO: Brace expansion

	for _, r := range c.cli.StringSlice("g5k-reserve-nodes") {
		// extract site name and number of nodes to reserve
		v := strings.Split(r, ":")

		// we only need 2 parameters : site and number of nodes
		if len(v) != 2 {
			return fmt.Errorf("Syntax error in nodes reservation parameter: '%s'", r)
		}

		// convert nodes number to int
		nb, err := strconv.Atoi(v[1])
		if err != nil {
			return fmt.Errorf("Error while converting number of nodes in reservation parameters: '%s'", r)
		}

		// store nodes to reserve for site
		c.nodesReservation[v[0]] = nb
	}

	return nil
}

// parseSwarmMasterFlag parse the Swarm Master flag (site)-(id)
func (c *CreateClusterCommand) parseSwarmMasterFlag() error {
	// initialize Swarm masters map
	c.swarmMasterNodes = make(map[string]bool)

	// TODO: Brace expansion

	for _, n := range c.cli.StringSlice("swarm-master") {
		// TODO: check if node exist (id too low/high)
		c.swarmMasterNodes[n] = true
	}

	return nil
}

// checkCliParameters perform checks on CLI parameters
func (c *CreateClusterCommand) checkCliParameters() error {
	// check username
	if c.cli.String("g5k-username") == "" {
		return fmt.Errorf("You must provide your Grid5000 account username")
	}

	// check password
	if c.cli.String("g5k-password") == "" {
		return fmt.Errorf("You must provide your Grid5000 account password")
	}

	// check nodes reservation
	if len(c.cli.StringSlice("g5k-reserve-nodes")) < 1 {
		return fmt.Errorf("You must provide a site and the number of nodes to reserve on it")
	}

	// parse nodes reservation
	if err := c.parseReserveNodesFlag(); err != nil {
		return err
	}

	// check walltime
	if c.cli.String("g5k-image") == "" {
		return fmt.Errorf("You must provide an image to deploy on the nodes")
	}

	// check walltime
	if c.cli.String("g5k-walltime") == "" {
		return fmt.Errorf("You must provide a walltime")
	}

	// check Swarm Standalone parameters
	if c.cli.Bool("swarm-standalone-enable") {
		// check Docker Swarm image
		if c.cli.String("swarm-standalone-image") == "" {
			return fmt.Errorf("You must provide a Swarm image")
		}

		// check Docker Swarm strategy
		if c.cli.String("swarm-standalone-strategy") == "" {
			return fmt.Errorf("You must provide a Swarm strategy")
		}
	}

	// check Swarm Mode parameters
	if c.cli.Bool("swarm-mode-enable") {
		// block enabling Swarm mode and Swarm standalone at the same time
		if c.cli.Bool("swarm-standalone-enable") {
			return fmt.Errorf("You can't enable both swarm modes at the same time")
		}

		// block enabling Weave Networking (unsupported with Swarm Mode)
		if c.cli.Bool("weave-networking") {
			return fmt.Errorf("You can't enable Weave networking with Swarm Mode (Only Swarm Standalone is supported)")
		}
	}

	// parse Swarm master flag only if Swarm is enabled
	if c.cli.Bool("swarm-standalone-enable") || c.cli.Bool("swarm-mode-enable") {
		if err := c.parseSwarmMasterFlag(); err != nil {
			return err
		}
	}

	return nil
}

// generateNodesGlobalConfig generate a Nodes global configuration from CLI flags
func (c *CreateClusterCommand) generateNodesGlobalConfig() error {
	// create nodes global configuration
	gc := &node.GlobalConfig{
		LibMachineClient:       libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir()),
		G5kUsername:            c.cli.String("g5k-username"),
		G5kPassword:            c.cli.String("g5k-password"),
		G5kImage:               c.cli.String("g5k-image"),
		G5kWalltime:            c.cli.String("g5k-walltime"),
		WeaveNetworkingEnabled: c.cli.Bool("weave-networking"),
		SwarmMasterNode:        c.swarmMasterNodes,
	}

	// Swarm Standalone config
	if c.cli.Bool("swarm-standalone-enable") {
		// enable Swarm Standalone
		gc.SwarmStandaloneGlobalConfig = &swarm.SwarmStandaloneGlobalConfig{
			Image:       c.cli.String("swarm-standalone-image"),
			Discovery:   c.cli.String("swarm-standalone-discovery"),
			Strategy:    c.cli.String("swarm-standalone-strategy"),
			MasterFlags: c.cli.StringSlice("swarm-standalone-opt"),
			JoinFlags:   c.cli.StringSlice("swarm-standalone-join-opt"),
		}

		// check Swarm discovery
		if c.cli.String("swarm-standalone-discovery") == "" {
			// generate new discovery token from Docker Hub
			if err := gc.SwarmStandaloneGlobalConfig.GenerateDiscoveryToken(); err != nil {
				return err
			}

			log.Infof("New Swarm discovery token generated : '%s'", gc.SwarmStandaloneGlobalConfig.Discovery)
		}
	}

	// enable Swarm Mode
	if c.cli.Bool("swarm-mode-enable") {
		gc.SwarmModeGlobalConfig = &swarm.SwarmModeGlobalConfig{}
	}

	// generate SSH key pair
	if err := gc.GenerateSSHKeyPair(); err != nil {
		return err
	}

	c.nodesGlobalConfig = gc
	return nil
}

// generateNodesConfig generate Nodes configuration from CLI flags
func (c *CreateClusterCommand) generateNodesConfig(site string, jobID int, deployedNodes []string) (map[string]node.Node, error) {
	// stores all deployed nodes configuration
	nodesConfig := make(map[string]node.Node)

	// create configuration for deployed nodes
	for i, n := range deployedNodes {
		// generate machine name : {site}-{id}
		machineName := fmt.Sprintf("%s-%d", site, i)

		// store node configuration
		nodesConfig[machineName] = node.Node{
			GlobalConfig: c.nodesGlobalConfig,
			NodeName:     n,
			MachineName:  machineName,
			G5kSite:      site,
			G5kJobID:     jobID,
		}
	}

	return nodesConfig, nil
}

// ProvisionNodes provision the nodes
func (c *CreateClusterCommand) provisionNodes(sites map[string]map[string]node.Node) error {
	log.Info("Starting nodes provisionning...")

	// TODO: bootstrap node
	/*
		- get a Swarm master node
		- provision the node
		- remove the node from the list
	*/

	// provision all deployed nodes
	var wg sync.WaitGroup
	for _, site := range sites {
		for _, n := range site {
			wg.Add(1)
			go func(n node.Node) {
				defer wg.Done()
				n.ProvisionNode()
			}(n)
		}
	}

	// wait nodes provisionning to finish
	wg.Wait()

	return nil
}

// CreateCluster create nodes in docker-machine
func (c *CreateClusterCommand) createCluster() error {
	// create Grid5000 API client
	g5kAPI := g5k.Init(c.cli.String("g5k-username"), c.cli.String("g5k-password"))

	// generate node global configuration
	if err := c.generateNodesGlobalConfig(); err != nil {
		return err
	}
	defer c.nodesGlobalConfig.LibMachineClient.Close()

	// stores the deployed nodes configuration by sites
	nodesConfigBySites := make(map[string]map[string]node.Node)

	// process nodes reservations by sites
	for site, nb := range c.nodesReservation {
		log.Infof("Reserving %d nodes on '%s' site...", nb, site)

		// reserve nodes
		jobID, err := g5kAPI.ReserveNodes(site, nb, c.cli.String("g5k-resource-properties"), c.cli.String("g5k-walltime"))
		if err != nil {
			return err
		}

		// deploy nodes
		deployedNodes, err := g5kAPI.DeployNodes(site, string(c.nodesGlobalConfig.SSHKeyPair.PublicKey), jobID, c.cli.String("g5k-image"))
		if err != nil {
			return err
		}

		// generate deployed nodes configuration
		nodesConfig, err := c.generateNodesConfig(site, jobID, deployedNodes)
		if err != nil {
			return err
		}

		// copy nodes configuration
		nodesConfigBySites[site] = nodesConfig
	}

	// provision deployed nodes
	if err := c.provisionNodes(nodesConfigBySites); err != nil {
		return err
	}

	return nil
}

// RunCreateClusterCommand create a new cluster using cli flags
func RunCreateClusterCommand(cli *cli.Context) error {
	c := CreateClusterCommand{cli: cli}

	// check CLI parameters
	if err := c.checkCliParameters(); err != nil {
		return err
	}

	// create the cluster
	if err := c.createCluster(); err != nil {
		return err
	}

	return nil
}
