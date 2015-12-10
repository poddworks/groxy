package main

import (
	"github.com/jeffjen/go-proxy/proxy"

	cli "github.com/codegangsta/cli"
	ctx "golang.org/x/net/context"

	"fmt"
	"os"
	"os/signal"
)

func main() {
	app := cli.NewApp()
	app.Name = "go-proxy"
	app.Usage = "The TCP proxy with discovery service support"
	app.Authors = []cli.Author{
		cli.Author{"Yi-Hung Jen", "yihungjen@gmail.com"},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "net", Usage: "Network type", Value: "tcp4"},
		cli.StringSliceFlag{Name: "src", Usage: "Origin address to listen"},
		cli.StringSliceFlag{Name: "dst", Usage: "Target to proxy to"},
		cli.StringSliceFlag{Name: "dsc", Usage: "Discovery service endpoint"},
		cli.StringFlag{Name: "srv", Usage: "Service identity in discovery"},
		cli.BoolFlag{Name: "lb", Usage: "Weather we do load balance"},
		cli.StringFlag{Name: "loglevel", Usage: "Set debug level", Value: "INFO", EnvVar: "LOG_LEVEL"},
	}
	app.Action = Proxy
	app.Run(os.Args)
}

func Proxy(c *cli.Context) {
	var (
		Net = c.String("net")
		Dsc = c.StringSlice("dsc")

		Lb = c.Bool("lb")

		From = make([]string, 0)

		meta *info
	)

	proxy.LogLevel(c.String("loglevel"))

	trigger := make(chan os.Signal, 1)
	signal.Notify(trigger, os.Interrupt, os.Kill)

	wk, abort := ctx.WithCancel(ctx.Background())

	if Net == "" {
		fmt.Fprintln(os.Stderr, "missing required flag --net")
		os.Exit(1)
	}
	if from := c.StringSlice("src"); len(from) != 0 {
		for _, one_from := range from {
			From = append(From, one_from)
		}
	}
	if len(From) == 0 {
		fmt.Fprintln(os.Stderr, "missing required flag --src")
		os.Exit(1)
	}
	if dst := c.StringSlice("dst"); len(dst) != 0 {
		var To = make([]string, len(dst))
		for idx, one_dst := range dst {
			To[idx] = one_dst
		}
		if len(From) == 1 {
			meta = &info{Net: Net, From: From[0], To: To, Balance: Lb}
		} else {
			meta = &info{Net: Net, FromRange: From, To: To, Balance: Lb}
		}
	} else if srv := c.String("srv"); srv != "" {
		if len(From) == 1 {
			meta = &info{Net: Net, From: From[0], Service: srv, Endpoints: Dsc, Balance: Lb}
		} else {
			meta = &info{Net: Net, FromRange: From, Service: srv, Endpoints: Dsc, Balance: Lb}
		}
	}

	// launch proxy worker
	halt := listen(wk, meta)

	// Block until a signal is received.
	<-trigger
	abort()

	fmt.Println("waiting...")
	<-halt
	fmt.Println("leaving now")
}
