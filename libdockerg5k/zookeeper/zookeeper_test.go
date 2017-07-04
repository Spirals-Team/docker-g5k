package zookeeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateClusterStorageURLSingleMaster(t *testing.T) {
	masters := []string{"lille-0"}
	hostsLookup := map[string]string{"lille-0": "10.0.0.0"}
	url := GenerateClusterStorageURL(masters, hostsLookup)
	assert.Equal(t, "zk://10.0.0.0", url)
}

func TestGenerateClusterStorageURLMultiMaster(t *testing.T) {
	masters := []string{"lille-0", "sophia-1", "lyon-2"}
	hostsLookup := map[string]string{"lille-0": "10.0.0.0", "sophia-1": "10.1.1.1", "lyon-2": "10.2.2.2"}
	url := GenerateClusterStorageURL(masters, hostsLookup)
	assert.Equal(t, "zk://10.0.0.0,10.1.1.1,10.2.2.2", url)
}

func TestGenerateServerListSingleMaster(t *testing.T) {
	masters := []string{"lille-0"}
	srvList := generateServerList(masters)
	assert.Equal(t, "server.0=lille-0:2888:3888", srvList)
}

func TestGenerateServerListMultiMaster(t *testing.T) {
	masters := []string{"lille-0", "sophia-1", "lyon-2"}
	srvList := generateServerList(masters)
	assert.Equal(t, "server.0=lille-0:2888:3888 server.1=sophia-1:2888:3888 server.2=lyon-2:2888:3888", srvList)
}
