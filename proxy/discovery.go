package proxy

import (
	disc "github.com/jeffjen/go-discovery"

	log "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"
	ctx "golang.org/x/net/context"

	"path"
)

var (
	retry = &Backoff{}
)

func watchWorker(c ctx.Context, watcher etcd.Watcher, key string) <-chan bool {
	v := make(chan bool)
	go func() {
		evt, err := watcher.Next(c)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Debug("watch")
			retry.Delay()
			v <- false
		} else {
			retry.Reset()
			log.WithFields(log.Fields{"Action": evt.Action, "Key": evt.Node.Key}).Debug("key space event")
			if evt.Action == "set" || evt.Action == "expire" || evt.Action == "delete" {
				if key == path.Dir(evt.Node.Key) {
					v <- true
				} else {
					v <- false
				}
			} else {
				v <- false
			}
		}
	}()
	return v
}

func obtainWorker(o chan<- []string, d *DiscOptions) chan<- bool {
	order := make(chan bool, 8)
	go func() {
		for _ = range order {
			nodes, err := obtain(d)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Debug("watch")
				o <- nil
			} else {
				o <- nodes
			}
		}
	}()
	return order
}

func watch(c ctx.Context, d *DiscOptions) (output <-chan []string, stop <-chan struct{}) {
	o, s := make(chan []string), make(chan struct{})
	go func() {
		defer close(s)
		watcher, err := disc.NewWatcher(&disc.WatcherOptions{
			Config:     etcd.Config{Endpoints: d.Endpoints},
			Key:        d.Service,
			AfterIndex: d.AfterIndex,
			Recursive:  true,
		})
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Warning("watch")
			return
		}
		order := obtainWorker(o, d)
		defer close(order)
		for yay := true; yay; {
			v := watchWorker(c, watcher, d.Service)
			select {
			case <-c.Done():
				yay = false
			case expect, ok := <-v:
				if ok && expect {
					order <- true
				}
				yay = ok
			}
		}
	}()
	output, stop = o, s
	return
}

func obtain(d *DiscOptions) ([]string, error) {
	cfg := etcd.Config{Endpoints: d.Endpoints}
	kAPI, err := disc.NewKeysAPI(cfg)
	if err != nil {
		return nil, err
	}
	resp, err := kAPI.Get(ctx.Background(), d.Service, &etcd.GetOptions{
		Recursive: true,
	})
	if err != nil {
		return nil, err
	}
	to := make([]string, 0)
	for _, n := range resp.Node.Nodes {
		if !n.Dir {
			to = append(to, path.Base(n.Key))
		}
	}
	log.WithFields(log.Fields{"To": to, "Service": d.Service}).Info("candidate")
	return to, nil
}
