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
Commands:
	{{range .Commands}}{{.Name}}{{ "\t " }}{{.Usage}}
	{{end}}
`
	cli.CommandHelpTemplate = `Usage: {{.Name}} {{if .Flags}}[OPTIONS]{{else if .ArgsUsage}}[CONFIG_KEY]{{end}}

{{.Usage}}

Options:
	{{range .Flags}}{{.}}
	{{end}}
`
}
