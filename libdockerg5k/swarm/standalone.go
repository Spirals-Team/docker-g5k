package swarm

import (
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/swarm/discovery/token"
)

// SwarmStandaloneGlobalConfig contain Swarm standalone global configuration
type SwarmStandaloneGlobalConfig struct {
	Image          string
	Discovery      string
	Strategy       string
	MasterFlags    []string
	JoinFlags      []string
	IsExperimental bool
}

// CreateSwarmStandaloneNodeConfig returns a configured SwarmOptions for HostOptions struct
func (gc *SwarmStandaloneGlobalConfig) CreateSwarmStandaloneNodeConfig(nodeName string, isMaster bool, isWorker bool) *swarm.Options {
	return &swarm.Options{
		IsSwarm:            true,
		Image:              gc.Image,
		Agent:              isWorker,
		Master:             isMaster,
		Discovery:          gc.Discovery,
		Address:            nodeName,
		Host:               "tcp://0.0.0.0:3376",
		Strategy:           gc.Strategy,
		ArbitraryFlags:     gc.MasterFlags,
		ArbitraryJoinFlags: gc.JoinFlags,
		IsExperimental:     gc.IsExperimental,
	}
}

// GetNewSwarmStandaloneDiscoveryToken get a new Docker Swarm discovery token from Docker Hub
func GetNewSwarmStandaloneDiscoveryToken() (string, error) {
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
