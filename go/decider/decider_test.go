package decider

import (
	"testing"
)

func TestLoad(t *testing.T) {
	path = "testdata/config.txt"
	load(true)
	if CheckDecider("decider", false) != true {
		t.Fatal("Get wrong value for decider.")
	}
}

func TestLoadNonexist(t *testing.T) {
	path = "testdata/config.txt"
	load(true)
	if CheckDecider("non_exist_decider", false) != false {
		t.Fatal("Get wrong value for non-exist decider.")
	}
}

func TestLoadUpdate(t *testing.T) {
	SetMockMode(true)
	defer SetMockMode(false)

	Mock("decider", 100)
	if CheckDecider("decider", false) != true {
		t.Fatal("Get wrong value for decider.")
	}

	// Use another config file to flip the decider and verify
	Mock("decider", 0)
	if CheckDecider("decider", false) != false {
		t.Fatal("Get wrong value for decider after update.")
	}
}
