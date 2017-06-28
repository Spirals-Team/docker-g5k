package hostsmapping

import (
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestGenerateHostsEntriesSingleIpv4(t *testing.T) {
	hostsLookupTable := map[string]string{"lille-0": "1.2.3.4"}
	entries := generateHostsEntries(hostsLookupTable)
	assert.Equal(t, fmt.Sprintf("\n# docker-g5k:\n1.2.3.4\tlille-0\n"), entries)
}

func TestGenerateHostsEntriesSingleIpv6(t *testing.T) {
	hostsLookupTable := map[string]string{"lille-0": "2001:db8:85a3::8a2e:370:7334"}
	entries := generateHostsEntries(hostsLookupTable)
	assert.Equal(t, fmt.Sprintf("\n# docker-g5k:\n2001:db8:85a3::8a2e:370:7334\tlille-0\n"), entries)
}
