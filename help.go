package main

import (
	log "github.com/Sirupsen/logrus"
	cli "github.com/codegangsta/cli"

	"os"
)

func init() {
	var level = os.Getenv("LOG_LEVEL")
	switch level {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
		break
	case "INFO":
		log.SetLevel(log.InfoLevel)
		break
	case "WARNING":
		log.SetLevel(log.WarnLevel)
		break
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
		break
	case "FATAL":
		log.SetLevel(log.FatalLevel)
		break
	case "PANIC":
		log.SetLevel(log.PanicLevel)
		break
	default:
		log.SetLevel(log.InfoLevel)
		break
	}

	cli.AppHelpTemplate = `Usage: {{.Name}} PROXY_SPEC [PROXY_SPEC ...]

{{.Usage}}

Version: {{.Version}}

PROXY_SPEC
	EXAMPLE SPEC: {"net": "tcp", "src": ":16379", "dst": [":6379"]}
	              {"net": "tcp", "srv": "/srv/redis/staging"}
`
}
