package weave

import (
	"strings"

	"github.com/docker/machine/libmachine/host"
)

const (
	weaveNetPath = "/usr/local/bin/weave"
	weaveNetURL  = "https://git.io/weave"

	weaveDiscoveryPath = "/usr/local/bin/discovery"
	weaveDiscoveryURL  = "https://raw.githubusercontent.com/weaveworks/discovery/master/discovery"
)

// InstallWeaveNet install Weave Net on given host
func InstallWeaveNet(h *host.Host) error {
	// download Weave Net script
	if _, err := h.RunSSHCommand(strings.Join([]string{"curl", "-L", weaveNetURL, "-o", weaveNetPath}, " ")); err != nil {
		return err
	}

	// add exec mode to Weave Net script
	if _, err := h.RunSSHCommand(strings.Join([]string{"chmod", "a+x", weaveNetPath}, " ")); err != nil {
		return err
	}

	return nil
}

// RunWeaveNet run Weave Net on given host
func RunWeaveNet(h *host.Host) error {
	// run Weave Net (weave launch)
	if _, err := h.RunSSHCommand(strings.Join([]string{weaveNetPath, "launch"}, " ")); err != nil {
		return err
	}

	return nil
}

// InstallWeaveDiscovery install Weave Discovery on given host
func InstallWeaveDiscovery(h *host.Host) error {
	// download Weave Discovery script
	if _, err := h.RunSSHCommand(strings.Join([]string{"curl", "-L", weaveDiscoveryURL, "-o", weaveDiscoveryPath}, " ")); err != nil {
		return err
	}

	// add exec mode to Weave Net script
	if _, err := h.RunSSHCommand(strings.Join([]string{"chmod", "a+x", weaveDiscoveryPath}, " ")); err != nil {
		return err
	}

	return nil
}

// RunWeaveDiscovery run Weave Discovery on a host using the given Swarm Discovery method
func RunWeaveDiscovery(h *host.Host, swarmDiscovery string) error {
	// run Weave Discovery (discovery join --advertise-router $SWARM_DISCOVERY)
	if _, err := h.RunSSHCommand(strings.Join([]string{"IMAGE_VERSION=latest", weaveDiscoveryPath, "join", "--discovered-port=6783", "--advertise-router", swarmDiscovery}, " ")); err != nil {
		return err
	}

	return nil
}
