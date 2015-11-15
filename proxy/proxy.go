package proxy

import (
	log "github.com/Sirupsen/logrus"
	ctx "golang.org/x/net/context"

	"net"
	"time"
)

// ConnOptions defines how the proxy should behave
type ConnOptions struct {
	Net          string
	From         string
	To           []string
	Service      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// To takes a Context and ConnOptiions and begin listening for request to
// proxy.
// Review https://godoc.org/golang.org/x/net/context for understanding the
// control flow.
func To(c ctx.Context, opts *ConnOptions) {
	ln, err := net.Listen(opts.Net, opts.From)
	if err != nil {
		log.Warning(err)
		return
	}
	newConn, quit, stop := acceptWorker(ln) // spawn worker to handle
	defer func() { ln.Close(); close(quit); <-stop }()
	for yay := true; yay; {
		select {
		case conn := <-newConn:
			work, _ := ctx.WithCancel(c)
			go handleConn(work, &connOrder{
				conn,
				opts.Net,
				opts.To,
				opts.ReadTimeout,
				opts.WriteTimeout,
			})
		case <-c.Done():
			yay = false
		}
	}
}
