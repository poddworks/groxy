/*
Transport level proxy for the mondern day.

The root package is provided as a standalone proxy app for verifying implementation detail.

This project aims to produce a programmable proxy to facilitate Ambassador
pattern: https://docs.docker.com/engine/articles/ambassador_pattern_linking/

A full implementation for a configurable Ambassador daemon ambd https://github.com/jeffjen/ambd

	Usage: groxy [OPTIONS]

	The TCP proxy with discovery service support

	Options:
			--net "tcp4"                            Network type
			--src [--src option --src option]       Origin address to listen
			--dst [--dst option --dst option]       Target to proxy to
			--dsc [--dsc option --dsc option]       Discovery service endpoint
			--srv                                   Service identity in discovery
			--lb                                    Weather we do load balance
			--loglevel "INFO"                       Set debug level [$LOG_LEVEL]
			--help, -h                              show help
			--version, -v                           print the version

	Commands:
			tls-client       Setup client encrypt mode
			tls-server       Setup server encrypt mode
			help             Shows a list of commands or help for one command

Running with static candidates:
	groxy --src :16379 --dst 10.0.3.144:6379

Running with static candidates and round robin balance:
	groxy --src :16379 --lb \
		--dst 10.0.0.12:6379 --dst 10.0.1.123:6379

Running with discovery backend:
    groxy --dsc http://etcd0:2379 --dsc http://etcd1:2379 \
        --src :16379 \
        --srv /srv/redis/debug

Running in cluster mode:
	groxy --src :16379 --src :16378 \
		--dst 10.0.0.12:6379 --dst 10.0.1.123:6379

Add TLS encryption to your connection
    groxy tls-client --src :16379 --dst 10.0.3.144:6379 \
        --tlscertpath s3://devops.example.org/client-cert

    groxy tls-client --src :16379 --dst 10.0.3.144:6379 \
        --tlscertpath /path/to/client-cert

Setting up TLS proxy server
    groxy tls-server --src :6379 --dst 10.0.3.144:6379 \
        --tlscertpath s3://devops.example.org/server-cert

    groxy tls-server --src :6379 --dst 10.0.3.144:6379 \
        --tlscertpath /path/to/server-cert

*/
package main
