package main

import (
	"vitess.io/vitess/go/mysql/knoxauth"
	"vitess.io/vitess/go/vt/vtgate"
)

func init() {
	vtgate.RegisterPluginInitializer(func() { knoxauth.Init() })
}
