package main

// This plugin imports opentsdb to register the opentsdb stats backend.

import (
	_ "vitess.io/vitess/go/stats/opentsdb"
)
