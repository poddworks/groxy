package main

import (
	"github.com/jeffjen/go-proxy/proxy"

	log "github.com/Sirupsen/logrus"
	ctx "golang.org/x/net/context"

	"errors"
)

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
