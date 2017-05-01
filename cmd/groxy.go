package main

import (
	"github.com/poddworks/groxy/cli"

	clii "github.com/urfave/cli"

	"os"
)

func main() {
	app := clii.NewApp()
	cli.SetupAppMetaData(app)
	cli.SetupFlags(app)
	cli.SetupCommand(app)
	cli.SetupBeforeProcessor(app)
	cli.SetupMainCommand(app)
	app.Run(os.Args)
}
