package swarm

import (
	"fmt"
	"net"

	"strings"

	"github.com/docker/machine/libmachine/host"
)

// SwarmModeGlobalConfig contain Swarm Mode global configuration
type SwarmModeGlobalConfig struct {
	ManagerToken        string
	BootstrapManagerURL string
	WorkerToken         string
}

// IsSwarmModeClusterInitialized returns true if Swarm mode cluster is initialized (Manager/Worker tokens set), and false otherwise
func (gc *SwarmModeGlobalConfig) IsSwarmModeClusterInitialized() bool {
	return (gc.ManagerToken != "") && (gc.WorkerToken != "")
}

// InitSwarmModeCluster initialize a new Swarm mode cluster on the given host and returns the Manager/Worker join tokens
func (gc *SwarmModeGlobalConfig) InitSwarmModeCluster(h *host.Host) error {
	// check if Swarm mode cluster is already initialized
	if gc.IsSwarmModeClusterInitialized() {
		return fmt.Errorf("The Swarm Mode cluster is already initialized")
	}

	// init Swarm mode cluster
	_, err := h.RunSSHCommand("docker swarm init")
	if err != nil {
		return err
	}

	// get Manager join token
	managerToken, err := h.RunSSHCommand("docker swarm join-token -q manager")
	if err != nil {
		return err
	}

	// get Worker join token
	workerToken, err := h.RunSSHCommand("docker swarm join-token -q worker")
	if err != nil {
		return err
	}

	// get IP address of the host
	ip, err := h.Driver.GetIP()
	if err != nil {
		return err
	}

	// remove spaces/new lines at the begining/end of the tokens
	gc.ManagerToken = strings.TrimSpace(managerToken)
	gc.WorkerToken = strings.TrimSpace(workerToken)

	// set this host as bootstrap Swarm Manager
	gc.BootstrapManagerURL = fmt.Sprintf("%s", net.JoinHostPort(ip, "2377"))

	return nil
}

// JoinSwarmModeCluster makes the host join a Swarm mode cluster as Manager or Worker
func (gc *SwarmModeGlobalConfig) JoinSwarmModeCluster(host *host.Host, isManager bool) error {
	// by default, join as Worker
	token := gc.WorkerToken

	// if node is a manager, change the join token to Manager token
	if isManager {
		token = gc.ManagerToken
	}

	// run swarm join command
	if _, err := host.RunSSHCommand(fmt.Sprintf("docker swarm join --token %s %s", token, gc.BootstrapManagerURL)); err != nil {
		return err
	}

	return nil
}
