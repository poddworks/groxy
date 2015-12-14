package proxy_test

import (
	ctx "golang.org/x/net/context"
	"log"
	"proxy"
)

/*
Create a proxy that listens on 0.0.0.0:6379 and foward to two remote
host, balance connections
*/
func ExampleTo_static() {
	pxyOpts := &proxy.ConnOptions{
		Net:     "tcp4",
		From:    ":6379",
		To:      []string{"10.0.0.12:6379", "10.0.1.123:6379"},
		Balance: true,
	}

	context, cancel := ctx.WithCancel(ctx.Backgroud())
	defer cancel()

	err := proxy.To(context, pxyOpts)
	log.Warning(err)
}

/*
Create a proxy that listens on 0.0.0.0:27017 and foward to hosts
registered under service key /srv/mongo_router/debug, trying them in
order.

New set of hosts is obtained on nodes joining or leaving service
/srv/mongo_router/debug, followed by connection reset.

See coreos/etcd https://github.com/coreos/etcd for more information on
discovery backend
*/
func ExampleSrv_discovery() {
	pxyOpts := &proxy.ConnOptions{
		Net:  "tcp4",
		From: ":27017",
		Discovery: &proxy.DiscOptions{
			Service:   "/srv/mongo_router/debug",
			Endpoints: []string{"http://10.0.1.11:2379", "http://10.0.2.13:2379"},
		},
	}

	context, cancel := ctx.WithCancel(ctx.Backgroud())
	defer cancel()

	err := proxy.Srv(context, pxyOpts)
	log.Warning(err)
}

/*
Create a proxy that connects each source endpoint to each remote endpoint.
Each connection behaves in the same way as proxy.To, but invoking
cancel function aborts both.
*/
func ExampleClusterTo_cluster() {
	pxyOpts := &proxy.ConnOptions{
		Net:       "tcp4",
		FromRange: []string{":16379", ":16378"},
		To:        []string{"10.0.0.12:6379", "10.0.1.123:6379"},
		Balance:   true,
	}

	context, cancel := ctx.WithCancel(ctx.Backgroud())
	defer cancel()

	err := proxy.ClusterTo(context, pxyOpts)
	log.Warning(err)
}
