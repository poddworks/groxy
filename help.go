package main

import (
	cli "github.com/codegangsta/cli"
)

func init() {
	cli.AppHelpTemplate = `Usage: {{.Name}} [OPTIONS]

{{.Usage}}

Options:
	{{range .Flags}}{{.}}
	{{end}}

Balance request between two redis node (READ ONLY)
	{{.Name}} --src :16379 --dst 127.0.0.1:6379 --dst 127.0.0.1:6380 --lb

Proxy to targets by service key name
	{{.Name}} --src :27017 --srv /srv/mongodb/debug --dsc http://localhost:2379

Many to many proxy
	{{.Name}} --src :37017 --src :37018 --dst localhost:27017 --dst localhost:27018

`
}
