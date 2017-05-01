/*
Transport level proxy for the mondern day.

The root package is provided as a standalone proxy app for verifying implementation detail.

	NAME:
	   groxy - The TCP proxy with discovery service support

	USAGE:
	   groxy [global options] command [command options] [arguments...]

	VERSION:
	   0.2.6

	AUTHOR:
	   Yi-Hung Jen <yihungjen@gmail.com>

	COMMANDS:
	     tls-client  Setup client encrypt mode
	     tls-server  Setup server encrypt mode
	     help, h     Shows a list of commands or help for one command

	GLOBAL OPTIONS:
	   --net value       Network type (default: "tcp4")
	   --src value       Origin address to listen
	   --dst value       Target to proxy to
	   --dsc value       Discovery service endpoint
	   --srv value       Service identity in discovery
	   --lb              Weather we do load balance
	   --loglevel value  Set debug level (default: "INFO") [$LOG_LEVEL]
	   --help, -h        show help
	   --version, -v     print the version

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
