# groxy

[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/poddworks/groxy/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/poddworks/groxy/proxy?status.png)](https://godoc.org/github.com/poddworks/groxy/proxy)
[![Build Status](https://travis-ci.org/poddworks/groxy.svg)](https://travis-ci.org/poddworks/groxy)

A simple proxy implementation for the modern day

*groxy* strives to achieve the following goals
- To be not complicated
- Code is the docuementation
- Exist as a standalone program and a library for proxy
- Interface with discovery backend
- Tunnel your data with TLS connection.
- Be as lean as possible

### Import groxy as library
`go get github.com/jeffjen/groxy/proxy`

### Running groxy as a standalone program
- Running with static candidates:
    ```
    groxy --src :16379 --dst 10.0.3.144:6379
    ```

- Running with discovery backend:
    ```
    groxy --src :16379 --srv /srv/redis/debug \
        --dsc http://etcd0:2379 \
        --dsc http://etcd1:2379
    ```

- Add TLS encryption to your connection
    ```
    groxy tls-client --src :16379 --dst 10.0.3.144:6379 \
        --tlscertpath s3://devops.example.org/client-cert

    groxy tls-client --src :16379 --dst 10.0.3.144:6379 \
        --tlscertpath /path/to/client-cert
    ```

- Setting up TLS proxy server
    ```
    groxy tls-server --src :6379 --dst 10.0.3.144:6379 \
        --tlscertpath s3://devops.example.org/server-cert

    groxy tls-server --src :6379 --dst 10.0.3.144:6379 \
        --tlscertpath /path/to/server-cert
    ```

### Behavior
The proxy module does not attempt to perform reconnection to remote endpoint.

The discovery backend chosen is etcd https://github.com/coreos/etcd

Layout for the discovery backend should look like the following:
```
/srv/redis/debug
/srv/redis/debug/10.0.1.134:6379
/srv/redis/debug/10.0.2.15:6379
/srv/redis/debug/10.0.3.41:6379
```

When nodes' state in service `/srv/redis/debug` changes e.g. leaving or joining,
the proxy will attempt to obtain a new set of nodes, followed by a reset on
established connections.

Proxy behaviors can be divided into two modes: ordered, or round-robin.

In ordered mode:
- the first remote host is attempted
- then second
- until all were tried, the proxy declare connection failed.

In round-robin node
- connections are spread among available candidates.
- no ordered retry

### Documentaion
GoDoc available: https://godoc.org/github.com/poddworks/groxy

