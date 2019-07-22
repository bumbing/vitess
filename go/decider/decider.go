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
	data      map[string]int
	mu        sync.RWMutex
	path      string
)

func init() {
	path = filePath
	load()
	go loop(30 * time.Second)
}

func CheckDecider(decider string, defaultValue bool) bool {
	mu.RLock()
	defer mu.RUnlock()
	if val, ok := data[decider]; ok && val > 0 {
		return rand.Intn(100) < val
	}
	return defaultValue
}

func load() {
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
			load()
		}
	}
}
