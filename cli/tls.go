package cli

import (
	"github.com/poddworks/groxy/proxy"

	clii "github.com/urfave/cli"

	"fmt"
	"os"
	"path"
)

var (
	certflags = []clii.Flag{
		// Certificate filepath with required keys
		clii.StringFlag{Name: "tlscertpath", Usage: "Specify certificate path", EnvVar: "GO_PROXY_CERTPATH"},

		// CA, Cert, and Key filepath
		clii.StringFlag{Name: "tlscacert", Usage: "Trust certs signed only by this CA", EnvVar: "GO_PROXY_CA"},
		clii.StringFlag{Name: "tlskey", Usage: "Path to TLS key file", EnvVar: "GO_PROXY_KEY"},
		clii.StringFlag{Name: "tlscert", Usage: "Path to TLS certificate file", EnvVar: "GO_PROXY_CERT"},
	}
)

func preprocess(c *clii.Context) (tlscacert, tlscert, tlskey string) {
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

func newTlsClientCommand() clii.Command {
	return clii.Command{
		Name:  "tls-client",
		Usage: "Setup client encrypt mode",
		Flags: append(common, certflags...),
		Before: func(c *clii.Context) error {
			setLoglevel(c)
			tlscacert, tlscert, tlskey := preprocess(c)
			opts := proxy.CertOptions{
				CA:      tlscacert,
				TlsCert: tlscert,
				TlsKey:  tlskey,
			}
			loadCertificate = func(c *clii.Context) proxy.TLSConfig {
				cfg, err := proxy.LoadCertificate(opts)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				return proxy.TLSConfig{Client: cfg}
			}
			return nil
		},
		Action: runProxy,
	}
}

func newTlsServerCommand() clii.Command {
	return clii.Command{
		Name:  "tls-server",
		Usage: "Setup server encrypt mode",
		Flags: append(common, certflags...),
		Before: func(c *clii.Context) error {
			setLoglevel(c)
			tlscacert, tlscert, tlskey := preprocess(c)
			opts := proxy.CertOptions{
				CA:      tlscacert,
				TlsCert: tlscert,
				TlsKey:  tlskey,
				Server:  true,
			}
			loadCertificate = func(c *clii.Context) proxy.TLSConfig {
				cfg, err := proxy.LoadCertificate(opts)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				return proxy.TLSConfig{Server: cfg}
			}
			return nil
		},
		Action: runProxy,
	}
}
