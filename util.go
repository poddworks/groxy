package main

import (
	"github.com/jeffjen/go-proxy/proxy"

	log "github.com/Sirupsen/logrus"
	ctx "golang.org/x/net/context"
)

type info struct {
	Net       string   `json:"net"`
	From      string   `json:"src"`
	FromRange []string `json:"range"`

	// static assignment
	To []string `json:"dst,omitempty"`

	// read from discovery
	Endpoints []string `json:"dsc,omitempty"`
	Service   string   `json:"srv,omitempty"`

	// balance connection request by origins
	Balance bool `json:"robin"`
}

func listen(wk ctx.Context, meta *info) (halt <-chan struct{}) {
	ending := make(chan struct{}, 1)
	go func() {
		defer close(ending)

		var err error

		logger := log.WithFields(log.Fields{
			"Net":       meta.Net,
			"From":      meta.From,
			"FromRange": meta.FromRange,
			"To":        meta.To,
			"Endpoints": meta.Endpoints,
			"Service":   meta.Service,
			"Balance":   meta.Balance,
		})

		logger.Info("begin")
		if meta.Service != "" && len(meta.Endpoints) != 0 {
			opts := &proxy.ConnOptions{
				Net:     meta.Net,
				Balance: meta.Balance,
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
				Net:     meta.Net,
				To:      meta.To,
				Balance: meta.Balance,
			}
			if len(meta.FromRange) == 0 {
				opts.From = meta.From
				err = proxy.To(wk, opts)
			} else {
				opts.FromRange = meta.FromRange
				err = proxy.ClusterTo(wk, opts)
			}
		}
		logger.WithFields(log.Fields{"err": err}).Warning("end")
	}()
	return ending
}
