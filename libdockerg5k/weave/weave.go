package weave

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/host"
)

const (
	weaveNetURL       = "https://git.io/weave"
	weaveDiscoveryURL = "https://raw.githubusercontent.com/weaveworks/discovery/master/discovery"
)

var (
	weaveNetPath       = mcndirs.GetBaseDir() + "/weave-net"
	weaveDiscoveryPath = mcndirs.GetBaseDir() + "/weave-discovery"
)

// downloadFile download a script from a given URL and store it to the provided path (It will Replace/Create the file and chmod 755)
func downloadScript(url, path string) error {
	// create new file (create/replace file, and chmod 755)
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	// download script
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// put downloaded data to a file
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}

// InstallLocalWeaveNetScript download the Weave Net script to the local Docker Machine base directory
func InstallLocalWeaveNetScript() error {
	// download script
	if err := downloadScript(weaveNetURL, weaveNetPath); err != nil {
		return err
	}

	return nil
}

// InstallLocalWeaveDiscoveryScript download the Weave Discovery script to the local Docker Machine base directory
func InstallLocalWeaveDiscoveryScript() error {
	// download script
	if err := downloadScript(weaveDiscoveryURL, weaveDiscoveryPath); err != nil {
		return err
	}

	return nil
}

// RunWeaveNet run Weave Net on given host
func RunWeaveNet(h *host.Host) error {
	// execute "weave launch"
	cmd := exec.Command(weaveNetPath, "launch", "--http-addr=0.0.0.0:6784")

	// get Docker engine URL for remote host
	url, err := h.URL()
	if err != nil {
		return err
	}

	// configure Docker env variables
	cmd.Env = []string{
		"DOCKER_TLS_VERIFY=1",
		fmt.Sprintf("DOCKER_HOST=%s", url),
		fmt.Sprintf("DOCKER_CERT_PATH=%s", h.AuthOptions().StorePath),
	}

	// execute command
	err = cmd.Start()
	if err != nil {
		return err
	}

	// wait command to finish
	cmd.Wait()

	return nil
}

// RunWeaveDiscovery run Weave Discovery on a host using the given Swarm Discovery method
func RunWeaveDiscovery(h *host.Host, swarmDiscovery string) error {
	// create command "discovery join --discovered-port=6783 --advertise-router $SWARM_DISCOVERY_METHOD"
	cmd := exec.Command(weaveDiscoveryPath, "join", "--discovered-port=6783", "--advertise-router", swarmDiscovery)

	// get Docker engine URL for remote host
	url, err := h.URL()
	if err != nil {
		return err
	}

	// configure Docker and Weave discovery env variables
	cmd.Env = []string{
		// Weave discovery
		"IMAGE_VERSION=latest",

		// Docker
		"DOCKER_TLS_VERIFY=1",
		fmt.Sprintf("DOCKER_HOST=%s", url),
		fmt.Sprintf("DOCKER_CERT_PATH=%s", h.AuthOptions().StorePath),
	}

	// execute command
	err = cmd.Start()
	if err != nil {
		return err
	}

	// wait command to finish
	cmd.Wait()

	return nil
}
