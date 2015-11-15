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
		network, from, to := parse(uri)
		log.WithFields(log.Fields{"Net": network, "From": from, "To": to}).Info("begin")
		proxy.To(wk, &proxy.ConnOptions{
			Net:  network,
			From: from,
			To:   to,
		})
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
