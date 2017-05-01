package discovery

import (
	etcd "github.com/coreos/etcd/client"
)

func NewKeysAPI(cfg etcd.Config) (etcd.KeysAPI, error) {
	eCli, err := etcd.New(cfg)
	if err != nil {
		return nil, err
	}
	return etcd.NewKeysAPI(eCli), nil
}

type WatcherOptions struct {
	Config etcd.Config

	Key string

	AfterIndex uint64
	Recursive  bool
}

func NewWatcher(opts *WatcherOptions) (etcd.Watcher, error) {
	kAPI, err := NewKeysAPI(opts.Config)
	if err != nil {
		return nil, err
	}
	watcher := kAPI.Watcher(opts.Key, &etcd.WatcherOptions{
		AfterIndex: opts.AfterIndex,
		Recursive:  opts.Recursive,
	})
	return watcher, nil
}
