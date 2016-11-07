package swarm

import "github.com/docker/swarm/discovery/token"

// GetNewSwarmDiscoveryToken get a new Docker Swarm discovery token from Docker Hub
func GetNewSwarmDiscoveryToken() (string, error) {
	// init Discovery structure
	discovery := token.Discovery{}
	discovery.Initialize("token", 0, 0, nil)

	// get a new discovery token from Docker Hub
	swarmToken, err := discovery.CreateCluster()
	if err != nil {
		return "", err
	}

	return swarmToken, nil
}
