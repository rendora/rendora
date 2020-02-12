
gitVersion ?= $(shell git describe --always)

build:
	CGO_ENABLED=0 go build -ldflags "-X main.gitVersion=$(gitVersion)"

install:
	CGO_ENABLED=0 go install -ldflags "-X main.gitVersion=$(gitVersion)"

clean:
	rm -f rendora
