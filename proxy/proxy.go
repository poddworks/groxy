package proxy

import (
	log "github.com/Sirupsen/logrus"
	ctx "golang.org/x/net/context"

	"net"
	"time"
)

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

type ConnOptiions struct {
	Net          string
	From         string
	To           []string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func To(c ctx.Context, opts *ConnOptiions) {
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

type connOrder struct {
	src net.Conn
	net string
	to  []string
	rd  time.Duration
	wd  time.Duration
}

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
		// FIXME: we need to handle this
		panic("unable to connect")
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
