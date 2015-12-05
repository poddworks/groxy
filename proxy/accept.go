package proxy

import (
	log "github.com/Sirupsen/logrus"
	ctx "golang.org/x/net/context"

	"net"
)

// AcceptWorker takes a net.Listener and starts Accept
// In place spawn goroutine
// Accepted connections are reported to newConn
// expect stop channel to close when AcceptWorker terminates
func AcceptWorker(c ctx.Context, ln net.Listener) (newConn <-chan net.Conn, stop <-chan struct{}) {
	nc, s := make(chan net.Conn, 8), make(chan struct{})
	go func() {
		defer close(s)
		for yay := true; yay; {
			select {
			default:
				conn, err := ln.Accept()
				if err != nil {
					log.WithFields(log.Fields{"err": err}).Debug("accept")
				} else {
					nc <- conn
				}
			case <-c.Done():
				yay = false
			}
		}
	}()
	newConn, stop = nc, s
	return
}
