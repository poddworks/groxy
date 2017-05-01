.PHONY: all linux darwin

all: linux darwin

clean:
	rm -f groxy-Linux-* groxy-Darwin-*

linux:
	env CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o groxy-Linux-x86_64 ./cmd

darwin:
	env CGO_ENABLED=0 GOOS=darwin go build -a -installsuffix cgo -o groxy-Darwin-x86_64 ./cmd
