package zookeeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateClusterStorageURLSingleMaster(t *testing.T) {
	masters := []string{"lille-0"}
	url := GenerateClusterStorageURL(masters)
	assert.Equal(t, "zk://lille-0", url)
}

func TestGenerateClusterStorageURLMultiMaster(t *testing.T) {
	masters := []string{"lille-0", "sophia-1", "lyon-2"}
	url := GenerateClusterStorageURL(masters)
	assert.Equal(t, "zk://lille-0,sophia-1,lyon-2", url)
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
