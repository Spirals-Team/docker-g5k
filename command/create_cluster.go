package command

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
	"github.com/kujtimiihoxha/go-brace-expansion"

	"github.com/Spirals-Team/docker-g5k/libdockerg5k/g5k"
	"github.com/Spirals-Team/docker-g5k/libdockerg5k/node"
	"github.com/Spirals-Team/docker-g5k/libdockerg5k/swarm"
)

const (
	// regexNodeName match the site (nodeSite) and the ID (nodeID) of a node from its name (nodeName)
	regexNodeName = "(?P<nodeName>(?P<nodeSite>[[:alpha:]]+)-(?P<nodeID>[[:digit:]]+))"

	// regexReservation match the site (site) and the number of nodes (nbNodes) from a reservation
	regexReservation = "^(?P<site>[[:alpha:]]+):(?P<nbNodes>[[:digit:]]+)$"

	// regexNodeParamFlag match the node site/ID and the parameter (param, paramName, paramValue) from a CLI flag using the format : {nodeName}:paramName=paramValue
	regexNodeParamFlag = "^" + regexNodeName + ":(?P<param>(?P<paramName>[[:ascii:]]+)=(?P<paramValue>[[:ascii:]]+))$"
)

var (
	// CreateClusterCliCommand represent the CLI command "create-cluster" with its flags
	CreateClusterCliCommand = cli.Command{
		Name:   "create-cluster",
		Usage:  "Create a new Docker Swarm cluster on the Grid'5000 infrastructure",
		Action: RunCreateClusterCommand,
		Flags: []cli.Flag{
			cli.StringFlag{
				EnvVar: "G5K_USERNAME",
				Name:   "g5k-username",
				Usage:  "Your Grid5000 account username",
				Value:  "",
			},

			cli.StringFlag{
				EnvVar: "G5K_PASSWORD",
				Name:   "g5k-password",
				Usage:  "Your Grid5000 account password",
				Value:  "",
			},

			cli.StringSliceFlag{
				EnvVar: "G5K_RESERVE_NODES",
				Name:   "g5k-reserve-nodes",
				Usage:  "Reserve nodes on a site (ex: lille:24)",
			},

			cli.StringFlag{
				EnvVar: "G5K_WALLTIME",
				Name:   "g5k-walltime",
				Usage:  "Machine's lifetime (HH:MM:SS)",
				Value:  "1:00:00",
			},

			cli.StringFlag{
				EnvVar: "G5K_IMAGE",
				Name:   "g5k-image",
				Usage:  "Name of the image to deploy",
				Value:  "jessie-x64-min",
			},

			cli.StringFlag{
				EnvVar: "G5K_RESOURCE_PROPERTIES",
				Name:   "g5k-resource-properties",
				Usage:  "Resource selection with OAR properties (SQL format)",
				Value:  "",
			},

			cli.StringSliceFlag{
				EnvVar: "ENGINE_OPT",
				Name:   "engine-opt",
				Usage:  "Specify arbitrary flags to include on the selected node(s) engine (site-id:flag=value)",
			},

			cli.StringSliceFlag{
				EnvVar: "ENGINE_LABEL",
				Name:   "engine-label",
				Usage:  "Specify labels for the selected node(s) engine (site-id:labelname=labelvalue)",
			},

			cli.StringSliceFlag{
				EnvVar: "SWARM_MASTER",
				Name:   "swarm-master",
				Usage:  "Select node(s) to be promoted to Swarm Master(standalone)/Manager(Mode)",
			},

			cli.BoolFlag{
				EnvVar: "SWARM_MODE_ENABLE",
				Name:   "swarm-mode-enable",
				Usage:  "Create a Swarm mode cluster",
			},

			cli.BoolFlag{
				EnvVar: "SWARM_STANDALONE_ENABLE",
				Name:   "swarm-standalone-enable",
				Usage:  "Create a Swarm standalone cluster",
			},

			cli.StringFlag{
				EnvVar: "SWARM_STANDALONE_DISCOVERY",
				Name:   "swarm-standalone-discovery",
				Usage:  "Discovery service to use with Swarm",
				Value:  "",
			},

			cli.StringFlag{
				EnvVar: "SWARM_STANDALONE_IMAGE",
				Name:   "swarm-standalone-image",
				Usage:  "Specify Docker image to use for Swarm",
				Value:  "swarm:latest",
			},

			cli.StringFlag{
				EnvVar: "SWARM_STANDALONE_STRATEGY",
				Name:   "swarm-standalone-strategy",
				Usage:  "Define a default scheduling strategy for Swarm",
				Value:  "spread",
			},

			cli.StringSliceFlag{
				EnvVar: "SWARM_STANDALONE_OPT",
				Name:   "swarm-standalone-opt",
				Usage:  "Define arbitrary flags for Swarm master (can be provided multiple times)",
			},

			cli.StringSliceFlag{
				EnvVar: "SWARM_STANDALONE_JOIN_OPT",
				Name:   "swarm-standalone-join-opt",
				Usage:  "Define arbitrary flags for Swarm join (can be provided multiple times)",
			},

			cli.BoolFlag{
				EnvVar: "WEAVE_NETWORKING",
				Name:   "weave-networking",
				Usage:  "Use Weave for networking (Only if Swarm standalone is enabled)",
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
	nodesEngineLabel  map[string][]string
	nodesEngineOpt    map[string][]string
}

// parseReserveNodesFlag parse the nodes reservation flag (site):(number of nodes)
func (c *CreateClusterCommand) parseReserveNodesFlag() error {
	// initialize nodes reservation map
	c.nodesReservation = make(map[string]int)

	for _, paramValue := range c.cli.StringSlice("g5k-reserve-nodes") {
		// brace expansion support
		for _, r := range gobrex.Expand(paramValue) {
			// extract site name and number of nodes to reserve
			v, err := ParseCliFlag(regexReservation, r)
			if err != nil {
				return fmt.Errorf("Syntax error in nodes reservation parameter: '%s'", paramValue)
			}

			// convert nodes number to int
			nb, err := strconv.Atoi(v["nbNodes"])
			if err != nil {
				return fmt.Errorf("Error while converting number of nodes in reservation parameters: '%s'", r)
			}

			// store nodes to reserve for site
			c.nodesReservation[v["site"]] = nb
		}
	}

	return nil
}

// parseSwarmMasterFlag parse the Swarm Master flag (site)-(id)
func (c *CreateClusterCommand) parseSwarmMasterFlag() error {
	// initialize Swarm masters map
	c.swarmMasterNodes = make(map[string]bool)

	for _, paramValue := range c.cli.StringSlice("swarm-master") {
		// brace expansion support
		for _, n := range gobrex.Expand(paramValue) {
			// TODO: check if node exist (id too low/high)
			c.swarmMasterNodes[n] = true
		}
	}

	return nil
}

// parseEngineOptFlag parse the nodes Engine Opt flag {site}-{id}:optname=optvalue
func (c *CreateClusterCommand) parseEngineOptFlag() error {
	// initialize nodes Engine Opt map
	c.nodesEngineOpt = make(map[string][]string)

	for _, paramValue := range c.cli.StringSlice("engine-opt") {
		for _, f := range gobrex.Expand(paramValue) {
			// extract node name and parameter
			v, err := ParseCliFlag(regexNodeParamFlag, f)
			if err != nil {
				return fmt.Errorf("Syntax error in node Engine flag parameter: '%s'", paramValue)
			}

			// append the parameter to the node's parameter list
			c.nodesEngineOpt[v["nodeName"]] = append(c.nodesEngineOpt[v["nodeName"]], v["param"])
		}
	}

	return nil
}

// parseEngineLabelFlag parse the nodes Engine label flag {site}-{id}:flagname=flagvalue
func (c *CreateClusterCommand) parseEngineLabelFlag() error {
	// initialize nodes Engine Label map
	c.nodesEngineLabel = make(map[string][]string)

	for _, paramValue := range c.cli.StringSlice("engine-label") {
		for _, f := range gobrex.Expand(paramValue) {
			// extract node name and parameter
			v, err := ParseCliFlag(regexNodeParamFlag, f)
			if err != nil {
				return fmt.Errorf("Syntax error in node Engine flag parameter: '%s'", paramValue)
			}

			// append the label to the node's label list
			c.nodesEngineLabel[v["nodeName"]] = append(c.nodesEngineLabel[v["nodeName"]], v["param"])
		}
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

	// parse engine opt
	if err := c.parseEngineOptFlag(); err != nil {
		return err
	}

	// parse engine label
	if err := c.parseEngineLabelFlag(); err != nil {
		return err
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
			EngineOpt:    c.nodesEngineOpt[machineName],
			EngineLabel:  c.nodesEngineLabel[machineName],
		}
	}

	return nodesConfig, nil
}

// ProvisionNodes provision the nodes
func (c *CreateClusterCommand) provisionNodes(deployedNodes *map[string]node.Node) error {
	log.Info("Starting nodes provisionning...")

	// Swarm bootstrap node :
	// We need to deploy one Swarm Master before any other nodes to get the Manager/Worker tokens (Swarm Mode)
	for k := range c.swarmMasterNodes {
		// get a bootstrap node (random Swarm Master)
		bootstrapNode := (*deployedNodes)[k]
		log.Infof("Swarm bootstrap node is '%s' ('%s')", bootstrapNode.MachineName, bootstrapNode.NodeName)

		// provision the node
		bootstrapNode.ProvisionNode()

		// remove the node from the list
		delete(*deployedNodes, k)

		break
	}

	// provision all deployed nodes
	var wg sync.WaitGroup
	for _, n := range *deployedNodes {
		wg.Add(1)
		go func(n node.Node) {
			defer wg.Done()
			n.ProvisionNode()
		}(n)
	}

	// wait nodes provisionning to finish
	wg.Wait()

	return nil
}

// appendSiteDeployedNodesConfig append the site deployed nodes config to the global nodes config
func (c *CreateClusterCommand) appendSiteDeployedNodesConfig(globalNodesConfig *map[string]node.Node, siteNodesConfig *map[string]node.Node) {
	for k, v := range *siteNodesConfig {
		(*globalNodesConfig)[k] = v
	}
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
	deployedNodesConfig := make(map[string]node.Node)

	// process nodes reservations by sites
	for site, nb := range c.nodesReservation {
		log.Infof("Reserving %d nodes on '%s' site...", nb, site)

		// reserve nodes
		jobID, err := g5kAPI.ReserveNodes(site, nb, c.cli.String("g5k-resource-properties"), c.cli.String("g5k-walltime"))
		if err != nil {
			return err
		}

		// deploy nodes
		deployedNodesName, err := g5kAPI.DeployNodes(site, string(c.nodesGlobalConfig.SSHKeyPair.PublicKey), jobID, c.cli.String("g5k-image"))
		if err != nil {
			return err
		}

		// generate deployed nodes configuration
		siteDeployedNodesConfig, err := c.generateNodesConfig(site, jobID, deployedNodesName)
		if err != nil {
			return err
		}

		// copy nodes configuration
		c.appendSiteDeployedNodesConfig(&deployedNodesConfig, &siteDeployedNodesConfig)
	}

	// provision deployed nodes
	if err := c.provisionNodes(&deployedNodesConfig); err != nil {
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
