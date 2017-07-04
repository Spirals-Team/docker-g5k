package swarm

import (
	"github.com/docker/machine/libmachine/swarm"
)

// SwarmStandaloneGlobalConfig contain Swarm standalone global configuration
type SwarmStandaloneGlobalConfig struct {
	Image       string
	Discovery   string
	Strategy    string
	MasterFlags []string
	JoinFlags   []string
}

// CreateNodeConfig returns a configured SwarmOptions for HostOptions struct
func (gc *SwarmStandaloneGlobalConfig) CreateNodeConfig(nodeName string, isMaster bool, isWorker bool) *swarm.Options {
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
		IsExperimental:     false,
	}
}
