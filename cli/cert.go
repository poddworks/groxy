package cli

import (
	"github.com/poddworks/groxy/proxy"

	clii "github.com/urfave/cli"
)

type LoadCertFunc func(c *clii.Context) proxy.TLSConfig

func noop(c *clii.Context) proxy.TLSConfig {
	return proxy.TLSConfig{}
}

var (
	loadCertificate LoadCertFunc = noop
)
