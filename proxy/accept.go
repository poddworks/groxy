package proxy

import (
	log "github.com/Sirupsen/logrus"

	"net"
)

// acceptWorker takes a net.Listener and starts Accept
// In place spawn goroutine
// Accepted connections are reported to newConn
// close quit channel to end acceptWorker
// expect stop channel to close when acceptWorker terminates
func acceptWorker(ln net.Listener) (newConn <-chan net.Conn, quit chan<- struct{}, stop <-chan struct{}) {
	nc, q, s := make(chan net.Conn), make(chan struct{}), make(chan struct{})
	go func() {
		defer close(s)
		for yay := true; yay; {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-q:
					yay = false
				default:
					log.Warning(err) // TODO: fix this?
				}
			} else {
				nc <- conn
			}
		}
	}()
	newConn, quit, stop = nc, q, s
	return
}
