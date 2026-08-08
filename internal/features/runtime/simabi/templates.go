package simabi

import _ "embed"

//go:embed templates/simulator.go.tmpl
var simulatorGoTemplate string

//go:embed templates/bridge.c.tmpl
var simulatorCTemplate string
