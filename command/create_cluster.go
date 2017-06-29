package command

import (
	"fmt"
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
	"github.com/kujtimiihoxha/go-brace-expansion"

	"github.com/Spirals-Team/docker-g5k/libdockerg5k/cluster"
	"github.com/Spirals-Team/docker-g5k/libdockerg5k/g5k"
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
	cli *cli.Context
}

// parseReserveNodesFlag parse the nodes reservation flag (site):(number of nodes)
func (c *CreateClusterCommand) parseReserveNodesFlag(flag []string) (map[string]int, error) {
	// initialize nodes reservation map
	nodesReservation := make(map[string]int)

	for _, paramValue := range flag {
		// brace expansion support
		for _, r := range gobrex.Expand(paramValue) {
			// extract site name and number of nodes to reserve
			v, err := ParseCliFlag(regexReservation, r)
			if err != nil {
				return nil, fmt.Errorf("Syntax error in nodes reservation parameter: '%s'", paramValue)
			}

			// convert nodes number to int
			nb, err := strconv.Atoi(v["nbNodes"])
			if err != nil {
				return nil, fmt.Errorf("Error while converting number of nodes in reservation parameters: '%s'", r)
			}

			// store nodes to reserve for site
			nodesReservation[v["site"]] = nb
		}
	}

	return nodesReservation, nil
}

// parseSwarmMasterFlag parse the Swarm Master flag (site)-(id)
func (c *CreateClusterCommand) parseSwarmMasterFlag(flag []string) (map[string]bool, error) {
	// initialize Swarm masters map
	swarmMasterNodes := make(map[string]bool)

	for _, paramValue := range flag {
		// brace expansion support
		for _, n := range gobrex.Expand(paramValue) {
			// extract site and node ID
			v, err := ParseCliFlag(regexNodeName, n)
			if err != nil {
				return nil, fmt.Errorf("Syntax error in Swarm master parameter: '%s'", paramValue)
			}

			swarmMasterNodes[v["nodeName"]] = true
		}
	}

	return swarmMasterNodes, nil
}

// parseEngineOptFlag parse the nodes Engine Opt flag {site}-{id}:optname=optvalue
func (c *CreateClusterCommand) parseEngineOptFlag(flag []string) (map[string][]string, error) {
	// initialize nodes Engine Opt map
	nodesEngineOpt := make(map[string][]string)

	for _, paramValue := range flag {
		for _, f := range gobrex.Expand(paramValue) {
			// extract node name and parameter
			v, err := ParseCliFlag(regexNodeParamFlag, f)
			if err != nil {
				return nil, fmt.Errorf("Syntax error in node Engine flag parameter: '%s'", paramValue)
			}

			// append the parameter to the node's parameter list
			nodesEngineOpt[v["nodeName"]] = append(nodesEngineOpt[v["nodeName"]], v["param"])
		}
	}

	return nodesEngineOpt, nil
}

// parseEngineLabelFlag parse the nodes Engine label flag {site}-{id}:flagname=flagvalue
func (c *CreateClusterCommand) parseEngineLabelFlag(flag []string) (map[string][]string, error) {
	// initialize nodes Engine Label map
	nodesEngineLabel := make(map[string][]string)

	for _, paramValue := range flag {
		for _, f := range gobrex.Expand(paramValue) {
			// extract node name and parameter
			v, err := ParseCliFlag(regexNodeParamFlag, f)
			if err != nil {
				return nil, fmt.Errorf("Syntax error in node Engine flag parameter: '%s'", paramValue)
			}

			// append the label to the node's label list
			nodesEngineLabel[v["nodeName"]] = append(nodesEngineLabel[v["nodeName"]], v["param"])
		}
	}

	return nodesEngineLabel, nil
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

	// check walltime
	if c.cli.String("g5k-image") == "" {
		return fmt.Errorf("You must provide an image to deploy on the nodes")
	}

	// check walltime
	if c.cli.String("g5k-walltime") == "" {
		return fmt.Errorf("You must provide a walltime")
	}

	// check if a Swarm master is defined (only if Swarm is enabled)
	if c.cli.Bool("swarm-standalone-enable") || c.cli.Bool("swarm-mode-enable") {
		if len(c.cli.StringSlice("swarm-master")) == 0 {
			return fmt.Errorf("You need to select your Swarm master node(s)")
		}
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

	return nil
}

// generateClusterConfig generate a cluster configuration from cli parameters
func (c *CreateClusterCommand) configureCluster() (*cluster.GlobalConfig, error) {
	// create nodes global configuration
	clusterConfig := &cluster.GlobalConfig{
		LibMachineClient:       libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir()),
		G5kUsername:            c.cli.String("g5k-username"),
		G5kPassword:            c.cli.String("g5k-password"),
		G5kImage:               c.cli.String("g5k-image"),
		G5kWalltime:            c.cli.String("g5k-walltime"),
		WeaveNetworkingEnabled: c.cli.Bool("weave-networking"),
		SwarmMasterNode:        make(map[string]bool),
	}

	// Swarm Standalone config
	if c.cli.Bool("swarm-standalone-enable") {
		// enable Swarm Standalone
		clusterConfig.SwarmStandaloneGlobalConfig = &swarm.SwarmStandaloneGlobalConfig{
			Image:       c.cli.String("swarm-standalone-image"),
			Discovery:   c.cli.String("swarm-standalone-discovery"),
			Strategy:    c.cli.String("swarm-standalone-strategy"),
			MasterFlags: c.cli.StringSlice("swarm-standalone-opt"),
			JoinFlags:   c.cli.StringSlice("swarm-standalone-join-opt"),
		}
	}

	// enable Swarm Mode
	if c.cli.Bool("swarm-mode-enable") {
		clusterConfig.SwarmModeGlobalConfig = &swarm.SwarmModeGlobalConfig{}
	}

	// generate SSH key pair
	if err := clusterConfig.GenerateSSHKeyPair(); err != nil {
		return nil, fmt.Errorf("Error while generating cluster SSH key pair: '%s'", err)
	}

	return clusterConfig, nil
}

// CreateCluster create nodes in docker-machine
func (c *CreateClusterCommand) createCluster() error {
	// create Grid5000 API client
	g5kAPI := g5k.Init(c.cli.String("g5k-username"), c.cli.String("g5k-password"))

	// generate cluster configuration from cli flags
	clusterConfig, err := c.configureCluster()
	if err != nil {
		return err
	}

	// create new cluster
	cluster := cluster.NewCluster(clusterConfig)
	defer cluster.Config.LibMachineClient.Close()

	// parse nodes reservation
	nodesReservation, err := c.parseReserveNodesFlag(c.cli.StringSlice("g5k-reserve-nodes"))
	if err != nil {
		return err
	}

	// create nodes in the cluster
	cluster.CreateNodes(nodesReservation)

	// parse engine opt
	engineOpts, err := c.parseEngineOptFlag(c.cli.StringSlice("engine-opt"))
	if err != nil {
		return err
	}

	// apply engine options to nodes
	for node, opts := range engineOpts {
		if _, ok := cluster.Nodes[node]; !ok {
			return fmt.Errorf("The node '%s' does not exist", node)
		}

		cluster.Nodes[node].EngineOpt = append(cluster.Nodes[node].EngineOpt, opts...)
	}

	// parse engine label
	engineLabels, err := c.parseEngineLabelFlag(c.cli.StringSlice("engine-label"))
	if err != nil {
		return err
	}

	// apply engine labels to nodes
	for node, labels := range engineLabels {
		if _, ok := cluster.Nodes[node]; !ok {
			return fmt.Errorf("The node '%s' does not exist", node)
		}

		cluster.Nodes[node].EngineLabel = append(cluster.Nodes[node].EngineLabel, labels...)
	}

	// parse Swarm master flag
	swarmMaster, err := c.parseSwarmMasterFlag(c.cli.StringSlice("swarm-master"))
	if err != nil {
		return err
	}

	// store swarm master nodes
	for node := range swarmMaster {
		if _, ok := cluster.Nodes[node]; !ok {
			return fmt.Errorf("The node '%s' does not exist", node)
		}

		cluster.Config.SwarmMasterNode[node] = true
	}

	// process nodes reservations by sites
	for site, nb := range nodesReservation {
		log.Infof("Reserving %d nodes on '%s' site...", nb, site)

		// reserve nodes
		jobID, err := g5kAPI.ReserveNodes(site, nb, c.cli.String("g5k-resource-properties"), c.cli.String("g5k-walltime"))
		if err != nil {
			return fmt.Errorf("Job reservation for site '%s' failed: '%s'", site, err)
		}

		// deploy nodes
		deployedNodes, err := g5kAPI.DeployNodes(site, string(cluster.Config.SSHKeyPair.PublicKey), jobID, c.cli.String("g5k-image"))
		if err != nil {
			return fmt.Errorf("Nodes deployment for site '%s' failed: '%s'", site, err)
		}

		// allocate deployed nodes to machines
		if err := cluster.AllocateDeployedNodesToMachines(site, jobID, deployedNodes); err != nil {
			return fmt.Errorf("Unable to allocate deployed nodes to machines for site '%s' : '%s'", site, err)
		}
	}

	// provision deployed nodes
	if err := cluster.ProvisionNodes(); err != nil {
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
