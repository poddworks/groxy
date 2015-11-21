package proxy

import (
	log "github.com/Sirupsen/logrus"
	ctx "golang.org/x/net/context"

	"errors"
	"net"
	"os"
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
)

// ConnOptions defines how the proxy should behave
type ConnOptions struct {
	Net          string
	From         string
	To           []string
	Discovery    *DiscOptions
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DiscOptions struct {
	Service    string
	Endpoints  []string
	AfterIndex uint64
}

// To takes a Context and ConnOptiions and begin listening for request to
// proxy.
// Review https://godoc.org/golang.org/x/net/context for understanding the
// control flow.
func To(c ctx.Context, opts *ConnOptions) error {
	ln, err := net.Listen(opts.Net, opts.From)
	if err != nil {
		return err
	}
	newConn, astp := AcceptWorker(c, ln) // spawn Accepter
	defer func() { ln.Close(); <-astp }()
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
	return ErrProxyEnd
}

func Srv(c ctx.Context, opts *ConnOptions) error {

	if opts.Discovery == nil {
		panic("DiscOptions missing")
	}
	candidates, err := Obtain(opts.Discovery)
	if err != nil {
		return err
	} else {
		opts.To = candidates
		log.WithFields(log.Fields{"To": opts.To}).Info("candidate")
	}
	ln, err := net.Listen(opts.Net, opts.From)
	if err != nil {
		return err
	}
	newConn, astp := AcceptWorker(c, ln)       // spawn Accepter
	newNodes, wstp := Watch(c, opts.Discovery) // spawn Watcher
	defer func() { ln.Close(); _, _ = <-astp, <-wstp }()

	var connList = make([]ctx.CancelFunc, 0)
	for yay := true; yay; {
		select {
		case opts.To = <-newNodes:
			// TODO: memory efficient way of doing this?
			for _, abort := range connList {
				abort()
			}
			connList = make([]ctx.CancelFunc, 0)
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
	return ErrProxyEnd
}
