package main

import (
	"github.com/jeffjen/go-proxy/proxy"

	cli "github.com/codegangsta/cli"
	ctx "golang.org/x/net/context"

	"fmt"
	"os"
	"os/signal"
)

var (
	common = []cli.Flag{
		cli.StringFlag{Name: "net", Usage: "Network type", Value: "tcp4"},
		cli.StringSliceFlag{Name: "src", Usage: "Origin address to listen"},
		cli.StringSliceFlag{Name: "dst", Usage: "Target to proxy to"},
		cli.StringSliceFlag{Name: "dsc", Usage: "Discovery service endpoint"},
		cli.StringFlag{Name: "srv", Usage: "Service identity in discovery"},
		cli.BoolFlag{Name: "lb", Usage: "Weather we do load balance"},
		cli.StringFlag{Name: "loglevel", Usage: "Set debug level", Value: "INFO", EnvVar: "LOG_LEVEL"},
	}
)

type LoadCertFunc func(c *cli.Context) proxy.TLSConfig

func noop(c *cli.Context) proxy.TLSConfig {
	return proxy.TLSConfig{}
}

var (
	loadCertificate LoadCertFunc = noop
)

func SetLoglevel(c *cli.Context) error {
	proxy.LogLevel(c.String("loglevel"))
	return nil
}

func Proxy(c *cli.Context) {
	var (
		Net = c.String("net")
		Dsc = c.StringSlice("dsc")

		Lb = c.Bool("lb")

		From = make([]string, 0)

		cert = loadCertificate(c)
	)

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
	Opts := &proxy.ConnOptions{
		Net:       Net,
		Balance:   Lb,
		TLSConfig: cert,
	}
	if dst := c.StringSlice("dst"); len(dst) != 0 {
		var To = make([]string, len(dst))
		for idx, one_dst := range dst {
			To[idx] = one_dst
		}
		Opts.To = To
		if len(From) == 1 {
			Opts.From = From[0]
		} else {
			Opts.FromRange = From
		}
	} else if srv := c.String("srv"); srv != "" {
		Opts.Discovery = &proxy.DiscOptions{
			Service:   srv,
			Endpoints: Dsc,
		}
		if len(From) == 1 {
			Opts.From = From[0]
		} else {
			Opts.FromRange = From
		}
	}

	// launch proxy worker
	halt := listen(wk, Opts)

	// Block until a signal is received.
	<-trigger
	abort()

	fmt.Println("waiting...")
	<-halt
	fmt.Println("leaving now")
}
