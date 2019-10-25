package decider

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"

	"vitess.io/vitess/go/vt/log"
)

const filePath = "/var/config/config.manageddata.admin.decider"

var (
	data         map[string]int
	mu           sync.RWMutex
	path         = filePath
	useMockData  = false
	mockData     = map[string]int{}
	monitorValue = []string{"use_pin_lookup_vindex", "pinvindex_sample_check", "dark_read_probability"}
)

func init() {
	load(true)
	go loop(30 * time.Second)
}

func getValueOrDefault(data map[string]int, decider string, defaultValue bool) bool {
	// If the value exists and it's in valid range, we should always prioritize it over using the default one.
	if val, ok := data[decider]; ok && val >= 0 {
		return rand.Intn(100) < val
	}
	return defaultValue
}

func CheckDecider(decider string, defaultValue bool) bool {
	mu.RLock()
	defer mu.RUnlock()
	if useMockData {
		return getValueOrDefault(mockData, decider, defaultValue)
	} else {
		return getValueOrDefault(data, decider, defaultValue)
	}
}

// Only for test purpose, to switch decider between using mocked value or
// config file.
func SetMockMode(mock bool) {
	mu.Lock()
	defer mu.Unlock()
	useMockData = mock
}

// Only for test purpose.
// Set one decider's value, and return a func to set it back.
// Usage:
//   Set it and keep the value:        Mock("decider", 100)
//   Set it back after function call:  defer Mock("decider", 100)
func Mock(name string, value int) func() {
	mu.Lock()
	defer mu.Unlock()
	temp, exist := mockData[name]
	mockData[name] = value
	return func() {
		mu.Lock()
		defer mu.Unlock()
		if exist {
			mockData[name] = temp
		} else {
			delete(mockData, name)
		}
	}
}

func load(firstLoad bool) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
		return
	}

	temp := map[string]int{}
	err = json.Unmarshal(bytes, &temp)
	if err != nil {
		log.Error(err)
		return
	}

	for _, key := range monitorValue {
		if !firstLoad && data[key] != temp[key] {
			log.Infof("%v value changed from %d to %d", key, data[key], temp[key])
		}
	}

	mu.Lock()
	defer mu.Unlock()
	data = temp
}

func loop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	lastModTime := time.Time{}
	for range ticker.C {
		stat, err := os.Stat(path)
		if err != nil {
			log.Error("could not stat file in path watcher: %v", err)
			continue
		}
		if stat.ModTime().After(lastModTime) {
			lastModTime = stat.ModTime()
			load(false)
		}
	}
}
