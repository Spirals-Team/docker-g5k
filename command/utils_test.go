package command

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testString     = "res1,42./test/val-res_test4"
	expectedResult = map[string]string{
		"":      testString,
		"test1": "res1",
		"test2": "42",
		"test3": "/test/val",
		"test4": "res_test4",
	}
)

func TestParseCliFlagRegexWithoutNamedCapturingGroup(t *testing.T) {
	regex := "([[:alnum:]]+),([[:digit:]]+).([[:ascii:]]+)-([[:word:]]+)"
	_, err := ParseCliFlag(regex, testString)

	assert.Error(t, err)
}

func TestParseCliFlagRegexWithNamedCapturingGroupCorrect(t *testing.T) {
	regex := "(?P<test1>[[:alnum:]]+),(?P<test2>[[:digit:]]+).(?P<test3>[[:ascii:]]+)-(?P<test4>[[:word:]]+)"
	res, err := ParseCliFlag(regex, testString)

	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(res, expectedResult))
}

func TestParseCliFlagRegexWithNamedCapturingGroupEmpty(t *testing.T) {
	regex := "(?P<test1>[[:alnum:]]+),(?P<test2>[[:digit:]]+).(?P<test3>[[:ascii:]]+)-(?P<test4>[[:word:]]+)"
	_, err := ParseCliFlag(regex, "")

	assert.Error(t, err)
}
