package proxy

import (
	log "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"
	ctx "golang.org/x/net/context"

	"path"
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
		_, err := watcher.Next(c)
		if err != nil {
			log.Debug(err)
			close(v)
		} else {
			v <- true
		}
	}()
	return v
}

func doObtain(o chan<- []string, d *DiscOptions) {
	nodes, err := Obtain(d)
	if err != nil {
		log.Warning(err)
	} else {
		o <- nodes
	}
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
		for yay := true; yay; {
			v := doWatch(c, watcher)
			select {
			case <-c.Done():
				yay = false
			case _, ok := <-v:
				if ok {
					go doObtain(o, d)
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
	return to, nil
}
