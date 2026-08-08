package deviceprobe

import _ "embed"

//go:embed templates/application.go.tmpl
var applicationSourceTemplate string

//go:embed templates/bootstrap.c
var bootstrapSource string

//go:embed templates/conservative_bootstrap.c
var conservativeBootstrapSource string

//go:embed templates/playdate.json
var targetSource string

//go:embed templates/adapter.S
var adapterSource string
