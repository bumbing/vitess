package vttablet

import (
	"vitess.io/vitess/go/event"
	"vitess.io/vitess/go/vt/hook"
	"vitess.io/vitess/go/vt/log"
	"vitess.io/vitess/go/vt/vttablet/tabletmanager/events"
)

func init() {
	event.AddListener(func(e *events.StateChange) {
		old := e.OldTablet
		alias, ks, shard := old.Alias, old.Keyspace, old.Shard
		if err := hook.NewHook("vttablet_state_change", []string{"-t", alias.String(), "-k", ks, "-s", shard}).ExecuteOptional(); err != nil {
			log.Errorf("vttablet_state_change hook execution failed %s", err)
		}
	})
}
