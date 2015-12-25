package proxy

import (
	log "github.com/Sirupsen/logrus"

	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/url"
	"time"
)

const (
	MAX_BACKOFF_DELAY = 2 * time.Second
)

// Backoff provides stepped delay on each failed attempts.  Maximum delay time
// is capped off at 2 seconds.
type Backoff struct {
	attempts int64
}

func (b *Backoff) min(x, y time.Duration) time.Duration {
	if x < y {
		return x
	} else {
		return y
	}
}

// Delay marks this attempt failed, increments counter, and sleep for no more
// then 2 seconds in this goroutine
func (b *Backoff) Delay() {
	b.attempts = b.attempts + 1
	delay := b.min(time.Duration(b.attempts)*2*time.Millisecond, MAX_BACKOFF_DELAY)
	log.WithFields(log.Fields{"after": delay, "attempts": b.attempts}).Debug("delay")
	time.Sleep(delay)
}

// Reset clears failed attempt counter
func (b *Backoff) Reset() {
	log.WithFields(log.Fields{"attempts": b.attempts}).Debug("reset")
	b.attempts = 0
}

// Attempts reports current failed attempts
func (b *Backoff) Attempts() int64 {
	return b.attempts
}

func loadCertFromFile(ca, tlscert, tlskey string) ([]tls.Certificate, *x509.CertPool, error) {
	cert, err := tls.LoadX509KeyPair(tlscert, tlskey)
	if err != nil {
		return nil, nil, err
	}

	ca_data, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, nil, err
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(ca_data) {
		return nil, nil, errors.New("Unable to process CA chain data")
	}

	return []tls.Certificate{cert}, pool, nil
}

func getfile(uri string) (p string, e error) {
	u, err := url.Parse(uri)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Debug("getfile")
		return "", err
	}
	logger := log.WithFields(log.Fields{"type": u.Scheme})
	switch {
	case u.Scheme == "" || u.Scheme == "file":
		p, e = uri, nil
		break
	case u.Scheme == "s3" || u.Scheme == "s3":
		p, e = news3cli().get(u)
		break
	case u.Scheme == "http" || u.Scheme == "https":
		p, e = "", errors.New("http/https not supported")
		break
	default:
		p, e = "", errors.New("unexpected resource URI "+u.Scheme)
		break
	}
	logger.WithFields(log.Fields{"uri": p, "err": e}).Debug("getfile")
	return
}

func loadCertCommon(ca, tlscert, tlskey string) ([]tls.Certificate, *x509.CertPool, error) {
	cafp, err := getfile(ca)
	if err != nil {
		return nil, nil, err
	}
	tlscertfp, err := getfile(tlscert)
	if err != nil {
		return nil, nil, err
	}
	tlskeyfp, err := getfile(tlskey)
	if err != nil {
		return nil, nil, err
	}
	return loadCertFromFile(cafp, tlscertfp, tlskeyfp)
}

// CertOptions provides specification to path of certificate and whether this
// is for listening server or connecting client
type CertOptions struct {
	// Certificate path information
	CA      string
	TlsCert string
	TlsKey  string

	// Setup tls.Config so we are listening server, otherwise we are connecting
	// client
	Server bool
}

// LoadCertificate processes certificate resource for TLSConfg to consume.
// Reads CA cert chain, Private, and Public key, then returns tls.Config.
//
// Supported URI:
//		- Local file on disk (file://)
// 		- Amazon Web Services S3 object (s3://)
//
// If left unspecified, URI is treated as if its a file reating on disk
//
// The default authentication rule is to verify cert key pair.
func LoadCertificate(opts CertOptions) (*tls.Config, error) {
	certs, pool, err := loadCertCommon(opts.CA, opts.TlsCert, opts.TlsKey)
	if err != nil {
		return nil, err
	}
	cfg := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: certs,
	}
	if opts.Server {
		cfg.ClientCAs = pool
	} else {
		cfg.RootCAs = pool
	}
	return cfg, nil
}
