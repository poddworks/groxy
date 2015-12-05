package proxy

import (
	log "github.com/Sirupsen/logrus"
	ctx "golang.org/x/net/context"

	"errors"
	"net"
	"os"
	"sync"
	"time"
)

func init() {
	var level = os.Getenv("LOG_LEVEL")
	switch level {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
		break
	case "INFO":
		log.SetLevel(log.InfoLevel)
		break
	case "WARNING":
		log.SetLevel(log.WarnLevel)
		break
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
		break
	case "FATAL":
		log.SetLevel(log.FatalLevel)
		break
	case "PANIC":
		log.SetLevel(log.PanicLevel)
		break
	default:
		log.SetLevel(log.InfoLevel)
		break
	}
}

var (
	ErrProxyEnd = errors.New("proxy end")

	ErrClusterNodeMismatch = errors.New("Origin and target count mismatch")
)

// ConnOptions defines how the proxy should behave
type ConnOptions struct {
	Net          string
	From         string
	FromRange    []string
	To           []string
	Discovery    *DiscOptions
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Balance      bool
}

type DiscOptions struct {
	Service    string
	Endpoints  []string
	AfterIndex uint64
}

func runTo(newConn <-chan net.Conn, c ctx.Context, opts *ConnOptions) {
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

func balanceTo(newConn <-chan net.Conn, c ctx.Context, opts *ConnOptions) {
	for yay, r := true, 0; yay; r = (r + 1) % len(opts.To) {
		select {
		case conn := <-newConn:
			work, _ := ctx.WithCancel(c)
			go handleConn(work, &connOrder{
				conn,
				opts.Net,
				opts.To[r : r+1],
				opts.ReadTimeout,
				opts.WriteTimeout,
			})
		case <-c.Done():
			yay = false
		}
	}
}

// To takes a Context and ConnOptions and begin listening for request to
// proxy.
// To obtains origin candidates through static listing.
// Review https://godoc.org/golang.org/x/net/context for understanding the
// control flow.
func To(c ctx.Context, opts *ConnOptions) error {
	ln, err := net.Listen(opts.Net, opts.From)
	if err != nil {
		return err
	}
	newConn, astp := AcceptWorker(c, ln) // spawn Accepter
	defer func() { ln.Close(); <-astp }()
	if opts.Balance {
		balanceTo(newConn, c, opts)
	} else {
		runTo(newConn, c, opts)
	}
	return ErrProxyEnd
}

func runSrv(newConn <-chan net.Conn, newNodes <-chan []string, c ctx.Context, opts *ConnOptions) {
	var connList = make([]ctx.CancelFunc, 0)
	for yay := true; yay; {
		select {
		case nodes := <-newNodes:
			if nodes != nil {
				opts.To = nodes
				// TODO: memory efficient way of doing this?
				for _, abort := range connList {
					abort()
				}
				connList = make([]ctx.CancelFunc, 0)
			}
		case conn := <-newConn:
			if len(opts.To) == 0 {
				conn.Close() // close connection to avoid confusion
			} else {
				work, abort := ctx.WithCancel(c)
				go handleConn(work, &connOrder{
					conn,
					opts.Net,
					opts.To,
					opts.ReadTimeout,
					opts.WriteTimeout,
				})
				connList = append(connList, abort)
			}
		case <-c.Done():
			yay = false
		}
	}
}

func balacnceSrv(newConn <-chan net.Conn, newNodes <-chan []string, c ctx.Context, opts *ConnOptions) {
	var connList = make([]ctx.CancelFunc, 0)
	for yay, r := true, 0; yay; r = (r + 1) % len(opts.To) {
		select {
		case nodes := <-newNodes:
			if nodes != nil {
				opts.To = nodes
				// TODO: memory efficient way of doing this?
				for _, abort := range connList {
					abort()
				}
				connList = make([]ctx.CancelFunc, 0)
			}
		case conn := <-newConn:
			if len(opts.To) == 0 {
				conn.Close() // close connection to avoid confusion
			} else {
				work, abort := ctx.WithCancel(c)
				go handleConn(work, &connOrder{
					conn,
					opts.Net,
					opts.To[r : r+1],
					opts.ReadTimeout,
					opts.WriteTimeout,
				})
				connList = append(connList, abort)
			}
		case <-c.Done():
			yay = false
		}
	}
}

// Srv takes a Context and ConnOptions and begin listening for request to
// proxy.
// Srv obtains origin candidates through discovery service by key.  If the
// candidate list changes in discovery record, Srv will reject current
// connections and obtain new origin candidates.
// Review https://godoc.org/golang.org/x/net/context for understanding the
// control flow.
func Srv(c ctx.Context, opts *ConnOptions) error {
	if opts.Discovery == nil {
		panic("DiscOptions missing")
	}
	if candidates, err := Obtain(opts.Discovery); err != nil {
		return err
	} else {
		opts.To = candidates
	}
	ln, err := net.Listen(opts.Net, opts.From)
	if err != nil {
		return err
	}
	newConn, astp := AcceptWorker(c, ln)       // spawn Accepter
	newNodes, wstp := Watch(c, opts.Discovery) // spawn Watcher
	defer func() { ln.Close(); _, _ = <-astp, <-wstp }()

	if opts.Balance {
		balacnceSrv(newConn, newNodes, c, opts)
	} else {
		runSrv(newConn, newNodes, c, opts)
	}

	return ErrProxyEnd
}

func ClusterTo(c ctx.Context, opts *ConnOptions) error {
	if len(opts.FromRange) > len(opts.To) {
		return ErrClusterNodeMismatch
	}
	var wg sync.WaitGroup
	for idx, from := range opts.FromRange {
		wg.Add(1)
		go func(from, to string) {
			// FIXME: need to report and err out
			To(c, &ConnOptions{
				Net:          opts.Net,
				From:         from,
				To:           []string{to},
				ReadTimeout:  opts.ReadTimeout,
				WriteTimeout: opts.WriteTimeout,
			})
			wg.Done()
		}(from, opts.To[idx])
	}
	<-c.Done()
	wg.Wait()
	return ErrProxyEnd
}

func ClusterSrv(c ctx.Context, opts *ConnOptions) error {
	if opts.Discovery == nil {
		panic("DiscOptions missing")
	}
	if candidates, err := Obtain(opts.Discovery); err != nil {
		return err
	} else {
		opts.To = candidates
	}
	if len(opts.FromRange) > len(opts.To) {
		return ErrClusterNodeMismatch
	}

	newNodes, wstp := Watch(c, opts.Discovery) // spawn Watcher
	defer func() { <-wstp }()

	for yay := true; yay; {
		var wg sync.WaitGroup
		work, abort := ctx.WithCancel(c)
		for idx, from := range opts.FromRange {
			var to []string
			if idx+1 > len(opts.To) {
				log.Warning("candidate node less then required range")
			} else {
				to = append(to, opts.To[idx])
			}
			wg.Add(1)
			go func(from string, to []string) {
				// FIXME: need to report and err out
				To(work, &ConnOptions{
					Net:          opts.Net,
					From:         from,
					To:           to,
					ReadTimeout:  opts.ReadTimeout,
					WriteTimeout: opts.WriteTimeout,
				})
				log.Debug("leave")
				wg.Done()
			}(from, to)
		}
		for yelp := true; yelp; {
			select {
			case nodes := <-newNodes:
				if nodes != nil {
					opts.To = nodes
					abort()
					yelp = false
				}
			case <-c.Done():
				abort()
				yay, yelp = false, false
			}
		}
		wg.Wait()
	}

	return ErrProxyEnd
}
