package weave

/*
Procedure to deploy Weave Net and Discovery:

Net:
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v /proc:/hostproc -e PROCFS=/hostproc --privileged --net=host weaveworks/weaveexec --local launch-router
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v /proc:/hostproc -e PROCFS=/hostproc --privileged --net=host weaveworks/weaveexec --local launch-plugin

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
	// Launch Weave Net Router
	if _, err := h.RunSSHCommand("docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v /proc:/hostproc -e PROCFS=/hostproc --privileged --net=host weaveworks/weaveexec --local launch-router"); err != nil {
		return err
	}

	// Launch Weave Net network plugin
	if _, err := h.RunSSHCommand("docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v /proc:/hostproc -e PROCFS=/hostproc --privileged --net=host weaveworks/weaveexec --local launch-plugin"); err != nil {
		return err
	}

	return nil
}

// RunWeaveDiscovery run Weave Discovery on a host using the given Swarm Discovery method
func RunWeaveDiscovery(h *host.Host, swarmDiscovery string) error {
	// Launch Weave Discovery
	if _, err := h.RunSSHCommand(fmt.Sprintf("docker run -d --name weavediscovery --net=host weaveworks/weavediscovery %s", swarmDiscovery)); err != nil {
		return err
	}

	return nil
}
