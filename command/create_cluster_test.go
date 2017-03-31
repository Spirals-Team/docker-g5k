package command

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test ParseReserveNodes flag
func TestParseReserveNodesFlagEmpty(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseReserveNodesFlag([]string{})
	assert.NoError(t, err)
}

func TestParseReserveNodesFlagIncorrectSiteName(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseReserveNodesFlag([]string{"site1:10"})
	assert.Error(t, err)
}

func TestParseReserveNodesFlagIncorrectFormatSingleValue(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseReserveNodesFlag([]string{"test=30"})
	assert.Error(t, err)
}

func TestParseReserveNodesFlagIncorrectFormatMultipleValue(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseReserveNodesFlag([]string{"test:30", "incorrect=20", "testt:10"})
	assert.Error(t, err)
}

func TestParseReserveNodesFlagCorrectFormatSingleValue(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseReserveNodesFlag([]string{"test:10"})
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(c.nodesReservation, map[string]int{"test": 10}))
}

func TestParseReserveNodesFlagCorrectFormatMultipleValue(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseReserveNodesFlag([]string{"test:10", "testt:20", "testtt:30"})
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(c.nodesReservation, map[string]int{"test": 10, "testt": 20, "testtt": 30}))
}

// Test ParseSwarmMaster flag
func TestParseSwarmMasterFlagEmpty(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseSwarmMasterFlag([]string{})
	assert.NoError(t, err)
}

func TestParseSwarmMasterFlagIncorrectSingleNodeName(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseSwarmMasterFlag([]string{"test"})
	assert.Error(t, err)
}

func TestParseSwarmMasterFlagIncorrectMultipleNodeName(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseSwarmMasterFlag([]string{"test-1", "incorrect", "test-2"})
	assert.Error(t, err)
}

func TestParseSwarmMasterFlagCorrectSingleNodeName(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseSwarmMasterFlag([]string{"test-1"})
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(c.swarmMasterNodes, map[string]bool{"test-1": true}))
}

func TestParseSwarmMasterFlagCorrectMultipleNodeName(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseSwarmMasterFlag([]string{"test-1", "test-2", "test-3"})
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(c.swarmMasterNodes, map[string]bool{"test-1": true, "test-2": true, "test-3": true}))
}

// Test ParseEngineOpt flag
func TestParseEngineOptFlagEmpty(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseEngineOptFlag([]string{})
	assert.NoError(t, err)
}

func TestParseEngineOptFlagIncorrectNodeNameSingleOpt(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseEngineOptFlag([]string{"incorrect:key=val"})
	assert.Error(t, err)
}

func TestParseEngineOptFlagIncorrectNodeNameMultipleOpt(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseEngineOptFlag([]string{"site-1:key=val", "incorrect:key=val"})
	assert.Error(t, err)
}

func TestParseEngineOptFlagCorrectNodeNameSingleOpt(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseEngineOptFlag([]string{"site-1:key=val"})
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(c.nodesEngineOpt, map[string][]string{"site-1": []string{"key=val"}}))
}

func TestParseEngineOptFlagCorrectNodeNameMultipleOptSingleValue(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseEngineOptFlag([]string{"site-1:key=val", "site-2:key=val"})
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(c.nodesEngineOpt, map[string][]string{
		"site-1": []string{"key=val"},
		"site-2": []string{"key=val"},
	}))
}

func TestParseEngineOptFlagCorrectNodeNameMultipleOptMultipleValue(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseEngineOptFlag([]string{"site-1:key1=val1", "site-1:key2=val2", "site-2:key=val"})
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(c.nodesEngineOpt, map[string][]string{
		"site-1": []string{"key1=val1", "key2=val2"},
		"site-2": []string{"key=val"},
	}))
}

// Test ParseEngineLabel flag
func TestParseEngineLabelFlagEmpty(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseEngineLabelFlag([]string{})
	assert.NoError(t, err)
}

func TestParseEngineLabelFlagIncorrectNodeNameSingleOpt(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseEngineLabelFlag([]string{"incorrect:key=val"})
	assert.Error(t, err)
}

func TestParseEngineLabelFlagIncorrectNodeNameMultipleOpt(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseEngineLabelFlag([]string{"site-1:key=val", "incorrect:key=val"})
	assert.Error(t, err)
}

func TestParseEngineLabelFlagCorrectNodeNameSingleOpt(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseEngineLabelFlag([]string{"site-1:key=val"})
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(c.nodesEngineLabel, map[string][]string{"site-1": []string{"key=val"}}))
}

func TestParseEngineLabelFlagCorrectNodeNameMultipleOptSingleValue(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseEngineLabelFlag([]string{"site-1:key=val", "site-2:key=val"})
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(c.nodesEngineLabel, map[string][]string{
		"site-1": []string{"key=val"},
		"site-2": []string{"key=val"},
	}))
}

func TestParseEngineLabelFlagCorrectNodeNameMultipleOptMultipleValue(t *testing.T) {
	c := CreateClusterCommand{}
	err := c.parseEngineLabelFlag([]string{"site-1:key1=val1", "site-1:key2=val2", "site-2:key=val"})
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(c.nodesEngineLabel, map[string][]string{
		"site-1": []string{"key1=val1", "key2=val2"},
		"site-2": []string{"key=val"},
	}))
}
