FROM scratch
MAINTAINER YI-HUNG JEN <yihungjen@gmail.com>

COPY ca-certificates.crt /etc/ssl/certs/
COPY groxy-Linux-x86_64 /groxy
ENTRYPOINT ["/groxy"]
CMD ["--help"]
