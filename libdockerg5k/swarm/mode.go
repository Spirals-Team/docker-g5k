package swarm

import "github.com/docker/machine/libmachine/host"

// JoinMode define the join modes for Swarm
type JoinMode int

const (
	// JoinAsManager makes the node join the cluster as Swarm Manager
	JoinAsManager JoinMode = iota
	// JoinAsWorker makes the node join the cluster as Swarm Worker
	JoinAsWorker
)

// InitSwarmModeCluster initialize a new Swarm mode cluster on the given host and returns the Manager/Worker join tokens
func InitSwarmModeCluster(h *host.Host) (string, string, error) {
	// Launch Weave Net Router
	if _, err := h.RunSSHCommand("docker swarm init"); err != nil {
		return "", "", err
	}

	return "", "", nil
}

// JoinSwarmModecluster makes the host join a Swarm mode cluster as Manager or Worker
func JoinSwarmModecluster(host *host.Host, joinMode JoinMode, joinToken string) error {
	return nil
}
