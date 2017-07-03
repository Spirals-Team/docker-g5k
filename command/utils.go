package command

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/Spirals-Team/docker-machine-driver-g5k/driver"
)

// ParseCliFlag extract informations (using regex) from cli flags and returns a map (named capturing groups are required)
func ParseCliFlag(regex string, str string) (map[string]string, error) {
	// compile the regex
	re := regexp.MustCompile(regex)

	// get the capturing groups names
	sn := re.SubexpNames()

	// extract informations from input string
	m := re.FindStringSubmatch(str)

	// test if the number of extracted informations matches the number of capturing groups
	if len(m) != len(sn) {
		return nil, fmt.Errorf("Number of extracted informations don't match the number of capturing groups")
	}

	// construct results map with capturing groups name and values
	r := make(map[string]string)
	for i, n := range sn {
		r[n] = m[i]
	}

	// test if named capturing groups were used (otherwise there is only one element in the map)
	if len(r) != len(sn) {
		return nil, fmt.Errorf("The use of named capturing groups is required")
	}

	return r, nil
}

// GetG5kDriverConfig takes a raw driver and return a configured Driver structure
func GetG5kDriverConfig(rawDriver []byte) (*driver.Driver, error) {
	// unmarshal driver configuration
	var drv driver.Driver
	if err := json.Unmarshal(rawDriver, &drv); err != nil {
		return nil, err
	}

	return &drv, nil
}
