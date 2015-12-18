package main

import (
	"github.com/jeffjen/go-proxy/proxy"

	cli "github.com/codegangsta/cli"

	"fmt"
	"os"
	"path"
)

var (
	certflags = []cli.Flag{
		// Certificate filepath with required keys
		cli.StringFlag{Name: "tlscertpath", Usage: "Specify ", EnvVar: "GO_PROXY_CERTPATH"},

		// CA, Cert, and Key filepath
		cli.StringFlag{Name: "tlscacert", Usage: "Trust certs signed only by this CA", EnvVar: "GO_PROXY_CA"},
		cli.StringFlag{Name: "tlskey", Usage: "Path to TLS key file", EnvVar: "GO_PROXY_KEY"},
		cli.StringFlag{Name: "tlscert", Usage: "Path to TLS certificate file", EnvVar: "GO_PROXY_CERT"},
	}
)

func preprocess(c *cli.Context) (tlscacert, tlscert, tlskey string) {
	var (
		// certificate path with all necessary info
		tlscertpath = c.String("tlscertpath")
	)

	// individual certificate file path
	tlscacert = c.String("tlscacert")
	tlskey = c.String("tlskey")
	tlscert = c.String("tlscert")

	if tlscertpath != "" {
		tlscacert = path.Join(tlscertpath, "ca.pem")
		tlskey = path.Join(tlscertpath, "key.pem")
		tlscert = path.Join(tlscertpath, "cert.pem")
	}
	if tlscacert == "" && tlskey == "" && tlscert == "" {
		fmt.Fprintln(os.Stderr, "No certificate specified")
		os.Exit(1)
	}

	return
}

func NewTlsClientCommand() cli.Command {
	return cli.Command{
		Name:  "tls-client",
		Usage: "Setup client encrypt mode",
		Flags: append(common, certflags...),
		Before: func(c *cli.Context) error {
			tlscacert, tlscert, tlskey := preprocess(c)
			opts := proxy.CertOptions{
				CA:      tlscacert,
				TlsCert: tlscert,
				TlsKey:  tlskey,
			}
			loadCertificate = func(c *cli.Context) proxy.TLSConfig {
				cfg, err := proxy.LoadCertificate(opts)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				return proxy.TLSConfig{Client: cfg}
			}
			return nil
		},
		Action: Proxy,
	}
}

func NewTlsServerCommand() cli.Command {
	return cli.Command{
		Name:  "tls-server",
		Usage: "Setup server encrypt mode",
		Flags: append(common, certflags...),
		Before: func(c *cli.Context) error {
			tlscacert, tlscert, tlskey := preprocess(c)
			opts := proxy.CertOptions{
				CA:      tlscacert,
				TlsCert: tlscert,
				TlsKey:  tlskey,
				Server:  true,
			}
			loadCertificate = func(c *cli.Context) proxy.TLSConfig {
				cfg, err := proxy.LoadCertificate(opts)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				return proxy.TLSConfig{Server: cfg}
			}
			return nil
		},
		Action: Proxy,
	}
}
