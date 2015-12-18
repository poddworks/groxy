package main

import (
	cli "github.com/codegangsta/cli"

	"os"
)

const (
	EncryptSrc  = "server"
	EncryptDst  = "client"
	EncryptBoth = "both"
	EncryptNone = ""
)

func main() {
	app := cli.NewApp()
	app.Name = "go-proxy"
	app.Usage = "The TCP proxy with discovery service support"
	app.Authors = []cli.Author{
		cli.Author{"Yi-Hung Jen", "yihungjen@gmail.com"},
	}
	app.Flags = common
	app.Commands = []cli.Command{
		NewTlsClientCommand(),
		NewTlsServerCommand(),
	}
	app.Action = Proxy
	app.Run(os.Args)
}
