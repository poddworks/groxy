package cli

import (
	"github.com/poddworks/groxy/proxy"

	log "github.com/Sirupsen/logrus"
	clii "github.com/urfave/cli"

	ctx "context"
	"errors"
	"fmt"
	"os"
	"os/signal"
)

var (
	common = []clii.Flag{
		clii.StringFlag{Name: "net", Usage: "Network type", Value: "tcp4"},
		clii.StringSliceFlag{Name: "src", Usage: "Origin address to listen"},
		clii.StringSliceFlag{Name: "dst", Usage: "Target to proxy to"},
		clii.StringSliceFlag{Name: "dsc", Usage: "Discovery service endpoint"},
		clii.StringFlag{Name: "srv", Usage: "Service identity in discovery"},
		clii.BoolFlag{Name: "lb", Usage: "Weather we do load balance"},
		clii.StringFlag{Name: "loglevel", Usage: "Set debug level", Value: "INFO", EnvVar: "LOG_LEVEL"},
	}
)

func SetupAppMetaData(app *clii.App) {
	app.Name = "groxy"

	app.Usage = "The TCP proxy with discovery service support"

	app.Version = "0.2.6"

	app.Authors = []clii.Author{
		clii.Author{"Yi-Hung Jen", "yihungjen@gmail.com"},
	}
}

func SetupFlags(app *clii.App) {
	app.Flags = common
}

func SetupBeforeProcessor(app *clii.App) {
	app.Before = setLoglevel
}

func SetupCommand(app *clii.App) {
	app.Commands = []clii.Command{
		newTlsClientCommand(),
		newTlsServerCommand(),
	}
}

func SetupMainCommand(app *clii.App) {
	app.Action = runProxy
}

func setLoglevel(c *clii.Context) error {
	proxy.LogLevel(c.String("loglevel"))
	return nil
}

func runProxy(c *clii.Context) {
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

func listen(wk ctx.Context, opts *proxy.ConnOptions) (halt <-chan struct{}) {
	var err error

	ending := make(chan struct{}, 1)
	go func() {
		defer close(ending)
		logger := log.WithFields(log.Fields{
			"Net":     opts.Net,
			"To":      opts.To,
			"Balance": opts.Balance,
		})
		if len(opts.FromRange) > 0 {
			logger = logger.WithFields(log.Fields{
				"FromRange": opts.FromRange,
			})
		} else {
			logger = logger.WithFields(log.Fields{
				"From": opts.From,
			})
		}
		if opts.Discovery != nil {
			logger = logger.WithFields(log.Fields{
				"Endpoints": opts.Discovery.Endpoints,
				"Service":   opts.Discovery.Service,
			})
		}
		logger.Info("begin")
		if opts.Discovery != nil {
			if len(opts.FromRange) == 0 {
				err = proxy.Srv(wk, opts)
			} else {
				err = proxy.ClusterSrv(wk, opts)
			}
		} else if len(opts.To) != 0 {
			if len(opts.FromRange) == 0 {
				err = proxy.To(wk, opts)
			} else {
				err = proxy.ClusterTo(wk, opts)
			}
		} else {
			err = errors.New("Misconfigured connect options")
		}
		logger.WithFields(log.Fields{"err": err}).Warning("end")

	}()
	return ending
}
