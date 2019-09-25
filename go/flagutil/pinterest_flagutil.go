package flagutil

import (
	"sort"
	"strings"
	"time"
)

// StringDurationMapValue is a map[string]Duration flag. It accepts a
// comma-separated list of key value pairs, of the form key:value. The
// keys cannot contain colons.
type StringDurationMapValue map[string]time.Duration

// Set sets the value of this flag from parsing the given string.
func (value *StringDurationMapValue) Set(v string) error {
	dict := make(map[string]time.Duration)
	pairs := parseListWithEscapes(v, ',')
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		duration, err := time.ParseDuration(parts[1])
		if err != nil {
			return err
		}
		dict[parts[0]] = duration
	}
	*value = dict
	return nil
}

// Get returns the map[string]time.Duration value of this flag.
func (value StringDurationMapValue) Get() interface{} {
	return map[string]time.Duration(value)
}

// String returns the string representation of this flag.
func (value StringDurationMapValue) String() string {
	parts := make([]string, 0)
	for k, v := range value {
		parts = append(parts, k+":"+v.String())
	}
	// Generate the string deterministically.
	sort.Strings(parts)
	return strings.Join(parts, ",")
}
