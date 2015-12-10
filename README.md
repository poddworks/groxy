# go-proxy
[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/jeffjen/go-libkv/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/jeffjen/go-proxy/proxy?status.png)](https://godoc.org/github.com/jeffjen/go-proxy/proxy)
[![Build Status](https://travis-ci.org/jeffjen/go-proxy.svg)](https://travis-ci.org/jeffjen/go-proxy)

A simple proxy implementation for the mondern day

This project aims to produce a programable proxy to facilitate the
**Ambassador** pattern.  Find out more about [running an Ambassador in micro
service deployment](https://github.com/jeffjen/ambd)

*go-proxy* strives to achieve the following goals
- To be not complicated
- Code is the docuementation
- Exist as a standalone program and a library for proxy
- Interface with discovery backend
- Be as lean as possible

### Import go-proxy as library
`go get github.com/jeffjen/go-proxy/proxy`

### Running go-proxy as a standalone program
- Running with static candidates:  
    `go-proxy --src :16379 --dst 10.0.3.144:6379`

- Running with discovery backend:  
    ```
    go-proxy --dsc http://etcd0:2379 --dsc http://etcd1:2379 \
        --src :16379 \
        --srv /srv/redis/debug
    ```

### Behavior
The proxy module does not attempt perform reconnection to remote endpoint.
The rule is fail fast and have the client retry.

The assumption here is that nodes are registered under key `/srv/redis/debug`,
with the last segment being the actual netloc of the node in this service
group i.e. `/srv/redis/debug/10.0.1.134:6379`, `/srv/redis/debug/10.0.2.122:6379`.

When the members in the `/srv/redis/debug` service changes e.g. leaving,
joining, the proxy will reject all established connections (if not already
broken from the view point of the client).

