package weave

/*
Procedure to deploy Weave Net and Discovery:

Net:
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v /proc:/hostproc -e PROCFS=/hostproc --privileged --net=host weaveworks/weaveexec --local launch-router --plugin

Discovery:
docker run -d --name weavediscovery --net=host weaveworks/weavediscovery $SWARM_DISCOVERY

Run containers with:
docker run --net=weave -h barbar.weave.local --dns=172.17.0.1 --dns-search=weave.local. -ti alpine:latest /bin/sh
*/

import (
	"fmt"

	"github.com/docker/machine/libmachine/host"
)

// RunWeaveNet run Weave Net on given host
func RunWeaveNet(h *host.Host) error {
	// Run Weave Net router with Docker plugin
	if _, err := h.RunSSHCommand("docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v /proc:/hostproc -e PROCFS=/hostproc --privileged --net=host weaveworks/weaveexec --local launch-router --plugin"); err != nil {
		return fmt.Errorf("Weave Net run command failed: '%s'", err)
	}

	return nil
}

// RunWeaveDiscovery run Weave Discovery on a host using the given Swarm Discovery method
func RunWeaveDiscovery(h *host.Host, swarmDiscovery string) error {
	// Run Weave Discovery
	if _, err := h.RunSSHCommand(fmt.Sprintf("docker run -d --name weavediscovery --net=host weaveworks/weavediscovery %s", swarmDiscovery)); err != nil {
		return fmt.Errorf("Weave Discovery run command failed: '%s'", err)
	}

	return nil
}
