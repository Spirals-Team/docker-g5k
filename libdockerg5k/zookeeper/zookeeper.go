package zookeeper

import (
	"fmt"
	"strings"

	"github.com/docker/machine/libmachine/host"
)

// GenerateClusterStorageURL returns a string used for Docker Engine/Swarm cluster-store parameter (format=zk://node1,node2,nodeN...)
func GenerateClusterStorageURL(zookeeperMasterNodes []string, hostsLookupTable map[string]string) string {
	// get the master nodes IP address from the hosts lookup table
	nodesIP := []string{}
	for _, n := range zookeeperMasterNodes {
		nodesIP = append(nodesIP, hostsLookupTable[n])
	}

	return fmt.Sprintf("zk://%s", strings.Join(nodesIP, ","))
}

// generateZookeeperServerList returns the list of zookeeper server for ZOO_SERVERS environment variable
func generateServerList(nodes []string) string {
	// generate zookeeper server string by nodes
	var zkServers []string
	for i, node := range nodes {
		zkServers = append(zkServers, fmt.Sprintf("server.%d=%s:2888:3888", i, node))
	}

	// returns the list as string
	return strings.Join(zkServers, " ")
}

// StartClusterStorage start a zookeeper k/vcontainer on the Swarm master nodes for cluster k/v storage
func StartClusterStorage(host *host.Host, zookeeperMasterNodes []string) error {
	// search current host in Swarm master nodes list
	for i, nodeName := range zookeeperMasterNodes {
		// host found in Swarm master nodes list
		if nodeName == host.Name {
			// construct needed zookeeper environment variables
			envID := fmt.Sprintf("ZOO_MY_ID=%d", i)
			envServers := fmt.Sprintf("ZOO_SERVERS=%s", generateServerList(zookeeperMasterNodes))

			// start zookeeper container
			if _, err := host.RunSSHCommand(fmt.Sprintf("docker run -td --restart=always --net=host --name docker-g5k-zookeeper -e \"%s\" -e \"%s\" zookeeper", envID, envServers)); err != nil {
				return err
			}

			return nil
		}
	}

	// host not found in Swarm master nodes list
	return fmt.Errorf("This host is not in the given Zookeeper master nodes list")
}
