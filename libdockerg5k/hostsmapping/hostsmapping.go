package hostsmapping

import (
	"bytes"
	"fmt"

	"github.com/docker/machine/libmachine/host"
)

// generateHostsEntries returns the entries as a single string
func generateHostsEntries(hostsLookupTable map[string]string) string {
	var buffer bytes.Buffer

	// append a header
	buffer.WriteString("\n# docker-g5k:\n")

	// entry format: {ip}<tab>{hostname}
	for hostname, ip := range hostsLookupTable {
		buffer.WriteString(fmt.Sprintf("%s\t%s\n", ip, hostname))
	}

	return buffer.String()
}

// AddClusterHostsMapping add cluster nodes name ({site}-{id}) to the static lookup table (/etc/hosts) of the node
func AddClusterHostsMapping(h *host.Host, hostsLookupTable map[string]string) error {
	// append entries at the end of the /etc/hosts file
	if _, err := h.RunSSHCommand(fmt.Sprintf("echo '%s' >>/etc/hosts", generateHostsEntries(hostsLookupTable))); err != nil {
		return fmt.Errorf("Failed to append hosts to the static lookup table: '%s'", err)
	}

	return nil
}
