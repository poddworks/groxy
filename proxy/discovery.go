package proxy

import (
	log "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"
	ctx "golang.org/x/net/context"

	"path"
)

var (
	retry Backoff
)

func NewKeysAPI(cfg etcd.Config) (etcd.KeysAPI, error) {
	eCli, err := etcd.New(cfg)
	if err != nil {
		return nil, err
	}
	return etcd.NewKeysAPI(eCli), nil
}

func NewWatcher(cfg etcd.Config, key string, index uint64) (etcd.Watcher, error) {
	kAPI, err := NewKeysAPI(cfg)
	if err != nil {
		return nil, err
	}
	watcher := kAPI.Watcher(key, &etcd.WatcherOptions{
		AfterIndex: index,
		Recursive:  true,
	})
	return watcher, nil
}

func doWatch(c ctx.Context, watcher etcd.Watcher) <-chan bool {
	v := make(chan bool)
	go func() {
		evt, err := watcher.Next(c)
		if err != nil {
			log.Debug(err)
			retry.Delay()
			v <- false
		} else {
			retry.Reset()
			log.WithFields(log.Fields{"Action": evt.Action, "Key": evt.Node.Key}).Debug("key space event")
			if evt.Action == "set" || evt.Action == "expire" || evt.Action == "delete" {
				v <- true
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
			nodes, err := Obtain(d)
			if err != nil {
				log.Warning(err)
				o <- nil
			} else {
				o <- nodes
			}
		}
	}()
	return order
}

func Watch(c ctx.Context, d *DiscOptions) (output <-chan []string, stop <-chan struct{}) {
	o, s := make(chan []string), make(chan struct{})
	go func() {
		defer close(s)
		cfg := etcd.Config{Endpoints: d.Endpoints}
		watcher, err := NewWatcher(cfg, d.Service, d.AfterIndex)
		if err != nil {
			log.Warning(err)
			return
		}
		order := obtainWorker(o, d)
		defer close(order)
		for yay := true; yay; {
			v := doWatch(c, watcher)
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

func Obtain(d *DiscOptions) ([]string, error) {
	cfg := etcd.Config{Endpoints: d.Endpoints}
	kAPI, err := NewKeysAPI(cfg)
	if err != nil {
		return nil, err
	}
	resp, err := kAPI.Get(ctx.Background(), d.Service, &etcd.GetOptions{
		Recursive: true,
	})
	if err != nil {
		return nil, err
	}
	to := make([]string, len(resp.Node.Nodes))
	for idx, n := range resp.Node.Nodes {
		to[idx] = path.Base(n.Key)
	}
	log.WithFields(log.Fields{"To": to}).Info("candidate")
	return to, nil
}
