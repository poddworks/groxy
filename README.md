# go-proxy
A simple proxy implementation

This project aims to produce a programable proxy to facilitate the
**Ambassador** pattern.  While there are many implementations for proxy out
there, this project hopes to achieve programable proxy in the easiest
representation possible.

*go-proxy* should achieve the following goals
- To be not complicated
- Code is the docuementation
- Exist as a standalone program and a library for proxy
- Interface with discovery backend
- Be as lean as possible

### Import go-proxy as library

`go get github.com/jeffjen/go-proxy/proxy`

### Running go-proxy as a standalone program

- Running with static candidates:  
    `LOG_LEVEL=DEBUG go-proxy '{"net": "tcp", "src": ":16379", "dst": [":6379"]}'`

- Running with discovery backend:  
    `LOG_LEVEL=DEBUG go-proxy '{"net": "tcp", "src": ":16379", "dsc": ["http://etcd0-ip:2379"], "srv": "/srv/redis/debug"}'`

The assumption here is that nodes are registered under key `/srv/redis/debug`,
with the last segment being the actual netloc of the node in this service
group i.e. `/srv/redis/debug/10.0.1.134:6379`, `/srv/redis/debug/10.0.2.122:6379`.

When the members in the `/srv/redis/debug` service changes e.g. leaving,
joining, the proxy will reject all established connections (if not already
broken from the view point of the client).  The proxy module does not attempt
perform reconnection to remote endpoint.  The rule is fail fast and have the
client retry.
