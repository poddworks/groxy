package proxy

import (
	log "github.com/Sirupsen/logrus"
	ctx "golang.org/x/net/context"

	"crypto/tls"
	"net"
	"reflect"
	"time"
)

// connOrder represents  a connection request
type connOrder struct {
	src net.Conn
	net string
	to  []string
	rd  time.Duration
	wd  time.Duration

	tlscfg *tls.Config
}

// handleConn establishes pipeline between the request party and the intended
// target to talk to.
func handleConn(c ctx.Context, work *connOrder) {
	var (
		src     net.Conn = work.src
		network string   = work.net
		to      []string = work.to

		dst net.Conn
		err error

		cert *tls.Config = work.tlscfg

		logger = log.WithFields(log.Fields{"-src": src.RemoteAddr(), "+src": src.LocalAddr(), "tls": cert != nil})
	)

	for _, addr := range to {
		if cert == nil {
			dst, err = net.Dial(network, addr)
		} else {
			dst, err = tls.Dial(network, addr, cert)
		}
		if err == nil {
			break
		}
	}
	if reflect.ValueOf(dst).IsNil() {
		logger.Debug("failed")
		src.Close()
		return
	} else {
		logger = logger.WithFields(log.Fields{"dst": dst.RemoteAddr()})
	}

	var (
		io  = NewCopyIO()
		one = make(chan struct{})
		two = make(chan struct{})
	)

	io.ReadDeadline(work.rd)
	io.WriteDeadline(work.wd)

	go func() {
		defer func() { close(one); dst.Close() }()
		io.Copy(dst, src)
	}()

	go func() {
		defer func() { close(two); src.Close() }()
		io.Copy(src, dst)
	}()

	logger.Debug("enter")
	for oy, ty := true, true; oy || ty; {
		select {
		case <-one:
			oy = false
		case <-two:
			ty = false
		case <-c.Done(): // force close on both ends
			src.Close()
			dst.Close()
		}
	}
	logger.Debug("leave")
}
