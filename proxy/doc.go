/*
Package proxy provides an intuitive transport level proxy.

The proxy module does not attempt to perform reconnection to remote endpoint.

The discovery backend chosen is etcd https://github.com/coreos/etcd

Layout for the discovery backend should look like the following:
	/srv/redis/debug
	/srv/redis/debug/10.0.1.134:6379
	/srv/redis/debug/10.0.2.15:6379
	/srv/redis/debug/10.0.3.41:6379

When nodes' state in service /srv/redis/debug changes e.g. leaving or joining,
the proxy will attempt to obtain a new set of nodes, followed by a reset on
established connections.

Proxy behaviors can be divided into two modes: ordered, or round-robin.

In ordered mode, the first remote host is attempted, then second, until all
were tried, the proxy declare connection failed.

In round-robin node, connections are spread among available candidates.
*/
package proxy
