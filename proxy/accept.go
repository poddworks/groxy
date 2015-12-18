package proxy

import (
	log "github.com/Sirupsen/logrus"
	ctx "golang.org/x/net/context"

	"crypto/tls"
	"net"
)

// config holds network type, address, and tls info
type config struct {
	network string
	addr    string
	tlscfg  *tls.Config
}

func newListener(c ctx.Context, cfg *config) (net.Listener, error) {
	var (
		logger = log.WithFields(log.Fields{"net": cfg.network, "from": cfg.addr, "tls": cfg.tlscfg != nil})

		retry = &Backoff{}
	)
	for {
		select {
		default:
			ln, err := net.Listen(cfg.network, cfg.addr)
			if err != nil {
				logger.WithFields(log.Fields{"err": err}).Warning("listen")
				retry.Delay()
			} else {
				// Upgrade connection to TLS if config available
				if cfg.tlscfg != nil {
					ln = tls.NewListener(ln, cfg.tlscfg)
				}
				logger.Debug("listen")
				return ln, nil
			}
		case <-c.Done():
			return nil, ErrProxyEnd
		}
	}
}

func accept(ln net.Listener) (conn <-chan net.Conn) {
	connChan := make(chan net.Conn)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Debug("accept")
		} else {
			connChan <- conn
		}
	}()
	return connChan
}

// acceptWorker takes a net.Listener and starts Accept
// In place spawn goroutine
// Accepted connections are reported to newConn
// expect stop channel to close when acceptWorker terminates
func acceptWorker(c ctx.Context, cfg *config) (newConn <-chan net.Conn, stop <-chan struct{}, err error) {
	ln, err := newListener(c, cfg)
	if err != nil {
		return
	}
	nc, s := make(chan net.Conn, 8), make(chan struct{})
	go func() {
		defer close(s)
		defer ln.Close()
		for yay := true; yay; {
			v := accept(ln)
			select {
			case <-c.Done():
				yay = false
			case conn, ok := <-v:
				if ok {
					nc <- conn
				}
				yay = ok
			}
		}
	}()
	newConn, stop, err = nc, s, nil
	return
}
