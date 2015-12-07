package proxy

import (
	log "github.com/Sirupsen/logrus"
	ctx "golang.org/x/net/context"

	"net"
)

func newListener(c ctx.Context, network, addr string) (net.Listener, error) {
	var retry = &Backoff{}
	for {
		select {
		default:
			ln, err := net.Listen(network, addr)
			if err != nil {
				log.WithFields(log.Fields{"net": network, "from": addr, "err": err}).Warning("listen")
				retry.Delay()
			} else {
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

// AcceptWorker takes a net.Listener and starts Accept
// In place spawn goroutine
// Accepted connections are reported to newConn
// expect stop channel to close when AcceptWorker terminates
func AcceptWorker(c ctx.Context, network, addr string) (newConn <-chan net.Conn, stop <-chan struct{}, err error) {
	ln, err := newListener(c, network, addr)
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
