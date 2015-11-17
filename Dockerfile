FROM scratch
MAINTAINER YI-HUNG JEN <yihungjen@gmail.com>

COPY ca-certificates.crt /etc/ssl/certs/
COPY go-proxy /
ENTRYPOINT ["/go-proxy"]
CMD ["--help"]
