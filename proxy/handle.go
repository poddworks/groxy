package proxy

import (
	log "github.com/Sirupsen/logrus"
	ctx "golang.org/x/net/context"

	"net"
	"time"
)

// connOrder represents  a connection request
type connOrder struct {
	src net.Conn
	net string
	to  []string
	rd  time.Duration
	wd  time.Duration
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
	)

	for _, addr := range to {
		dst, err = net.Dial(network, addr)
		if err == nil {
			break
		}
	}
	if dst == nil {
		log.WithFields(log.Fields{"+src": src.LocalAddr(), "-src": src.RemoteAddr()}).Debug("failed")
		src.Close()
		return
	}

	var (
		io = NewCopyIO()

		one = make(chan struct{})
		two = make(chan struct{})

		f = log.Fields{"+src": src.LocalAddr(), "-src": src.RemoteAddr(), "dst": dst.RemoteAddr()}
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

	log.WithFields(f).Debug("enter")
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
	log.WithFields(f).Debug("leave")
}
