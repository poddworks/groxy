package main

import (
	"github.com/jeffjen/go-proxy/proxy"

	log "github.com/Sirupsen/logrus"
	cli "github.com/codegangsta/cli"
	ctx "golang.org/x/net/context"

	"os"
	"os/signal"
	"sync"
)

func main() {
	app := cli.NewApp()
	app.Name = "go-proxy"
	app.Usage = "Facilitate TCP proxy"
	app.Authors = []cli.Author{
		cli.Author{"Yi-Hung Jen", "yihungjen@gmail.com"},
	}
	app.Action = Proxy
	app.Run(os.Args)
}

func listen(wg *sync.WaitGroup, uri string) ctx.CancelFunc {
	wk, abort := ctx.WithCancel(ctx.Background())
	go func() {
		defer wg.Done()

		meta, err := parse(uri)
		if err != nil {
			log.Warning(err)
			return
		}

		fields := log.Fields{"Net": meta.Net, "From": meta.From, "FromRange": meta.FromRange, "To": meta.To, "Endpoints": meta.Endpoints, "Service": meta.Service}

		log.WithFields(fields).Info("begin")
		if meta.Service != "" && len(meta.Endpoints) != 0 {
			opts := &proxy.ConnOptions{
				Net: meta.Net,
				Discovery: &proxy.DiscOptions{
					Service:   meta.Service,
					Endpoints: meta.Endpoints,
				},
			}
			if len(meta.FromRange) == 0 {
				opts.From = meta.From
				err = proxy.Srv(wk, opts)
			} else {
				opts.FromRange = meta.FromRange
				err = proxy.ClusterSrv(wk, opts)
			}
		} else if len(meta.To) != 0 {
			opts := &proxy.ConnOptions{
				Net: meta.Net,
				To:  meta.To,
			}
			if len(meta.FromRange) == 0 {
				opts.From = meta.From
				err = proxy.To(wk, opts)
			} else {
				opts.FromRange = meta.FromRange
				err = proxy.ClusterTo(wk, opts)
			}
		}
		log.WithFields(fields).Warning(err)
	}()
	return abort
}

func Proxy(c *cli.Context) {
	var (
		wg      sync.WaitGroup
		workers = make(map[string]ctx.CancelFunc)
	)

	trigger := make(chan os.Signal, 1)
	signal.Notify(trigger, os.Interrupt, os.Kill)

	for _, uri := range c.Args() {
		wg.Add(1)
		workers[uri] = listen(&wg, uri)
	}

	if len(workers) == 0 {
		log.Info("nothing to do, abort...")
		return
	}

	// Block until a signal is received.
	<-trigger
	for _, abort := range workers {
		abort()
	}

	// Reap all workers
	log.Info("waiting...")
	wg.Wait()
	log.Info("leaving now")
}
