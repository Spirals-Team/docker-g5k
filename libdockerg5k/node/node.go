package node

import (
	"github.com/Spirals-Team/docker-g5k/libdockerg5k/swarm"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/ssh"
)

// GlobalConfig contain global nodes configuration
type GlobalConfig struct {
	// libMachine client
	LibMachineClient *libmachine.Client

	// g5k driver config (needed or Docker Machine will not work afterwards)
	G5kUsername string
	G5kPassword string
	G5kImage    string
	G5kWalltime string
	SSHKeyPair  *ssh.KeyPair

	// Swarm configuration
	SwarmStandaloneGlobalConfig *swarm.SwarmStandaloneGlobalConfig
	SwarmModeGlobalConfig       *swarm.SwarmModeGlobalConfig
	SwarmMasterNode             map[string]bool

	// Weave networking
	WeaveNetworkingEnabled bool
}

// GenerateSSHKeyPair generate a new global SSH key pair
func (gc *GlobalConfig) GenerateSSHKeyPair() error {
	sshKeyPair, err := ssh.NewKeyPair()
	if err != nil {
		return err
	}

	gc.SSHKeyPair = sshKeyPair
	return nil
}

// Node contain node specific informations
type Node struct {
	*GlobalConfig

	NodeName    string // Grid'5000 node hostname
	MachineName string // Docker Machine name

	// g5k driver
	G5kSite  string
	G5kJobID int

	// Docker Engine
	EngineOpt   []string
	EngineLabel []string
}
