# go-proxy
A simple proxy implementation

This project aims to produce a programable proxy to facilitate the **Ambassador** pattern.  While there are many implementations for proxy out there, this project hopes to achieve programable proxy in the easiest representation possible.

*go-proxy* should achieve the following goals
- To be not complicated
- Code is the docuementation
- Exist as a standalone program and a library for proxy
- Interface with discovery backend
- Be as lean as possible

### Import go-proxy as library

`go get github.com/jeffjen/go-proxy/proxy`

### Running go-proxy as a standalone program

`LOG_LEVEL=DEBUG go-proxy '{"net": "tcp", "src": ":16379", "dst": [":6379"]}'`
