# go-proxy

[![Join the chat at https://gitter.im/jeffjen/go-proxy](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/jeffjen/go-proxy?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/jeffjen/go-libkv/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/jeffjen/go-proxy/proxy?status.png)](https://godoc.org/github.com/jeffjen/go-proxy/proxy)
[![Build Status](https://travis-ci.org/jeffjen/go-proxy.svg)](https://travis-ci.org/jeffjen/go-proxy)

A simple proxy implementation for the modern day

This project aims to produce a programmable proxy to facilitate **Ambassador**
pattern.  Find out more about [Ambassador in micro service
deployment](https://github.com/jeffjen/ambd)

*go-proxy* strives to achieve the following goals
- To be not complicated
- Code is the docuementation
- Exist as a standalone program and a library for proxy
- Interface with discovery backend
- Tunnel your data with TLS connection.
- Be as lean as possible

### Import go-proxy as library
`go get github.com/jeffjen/go-proxy/proxy`

### Running go-proxy as a standalone program
- Running with static candidates:
    ```
    go-proxy --src :16379 --dst 10.0.3.144:6379
    ```

- Running with discovery backend:
    ```
    go-proxy --src :16379 --srv /srv/redis/debug \
        --dsc http://etcd0:2379 \
        --dsc http://etcd1:2379
    ```

- Add TLS encryption to your connection
    ```
    go-proxy tls-client --src :16379 --dst 10.0.3.144:6379 \
        --tlscertpath s3://devops.example.org/client-cert

    go-proxy tls-client --src :16379 --dst 10.0.3.144:6379 \
        --tlscertpath /path/to/client-cert
    ```

- Setting up TLS proxy server
    ```
    go-proxy tls-server --src :6379 --dst 10.0.3.144:6379 \
        --tlscertpath s3://devops.example.org/server-cert

    go-proxy tls-server --src :6379 --dst 10.0.3.144:6379 \
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
GoDoc available: https://godoc.org/github.com/jeffjen/go-proxy

