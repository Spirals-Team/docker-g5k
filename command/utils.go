package command

import (
	"fmt"
	"regexp"
)

// ParseCliFlag extract informations (using regex) from cli flags and returns a map (named capturing groups are required)
func ParseCliFlag(regex string, str string) (map[string]string, error) {
	// compile the regex
	re := regexp.MustCompile(regex)

	// get the capturing groups names
	sn := re.SubexpNames()
	if len(sn) == 0 {
		return nil, fmt.Errorf("The use of named capturing groups is required")
	}

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

	return r, nil
}
