package main

import (
	cli "github.com/codegangsta/cli"
)

func init() {
	cli.AppHelpTemplate = `Usage: {{.Name}} PROXY_SPEC [PROXY_SPEC ...]

{{.Usage}}

Version: {{.Version}}

PROXY_SPEC
	EXAMPLE SPEC: {"net": "tcp", "src": ":16379", "dst": [":6379"]}
	              {"net": "tcp", "srv": "/srv/redis/staging"}
`
}
