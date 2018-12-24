
VERSION ?= $(shell git describe --always)
GOVERSION ?= $(shell go version)

PREFIX ?= /usr
BINPREFIX ?= $(PREFIX)/bin
PROGRAM := rendora

build:
	cd ./cmd/rendora; \
	CGO_ENABLED=0 go build -ldflags "-X main.VERSION=$(VERSION)"

install:
	mkdir -p "$(DESTDIR)$(BINPREFIX)"
	cp -pf cmd/rendora/$(PROGRAM) "$(DESTDIR)$(BINPREFIX)"

clean:
	rm -f cmd/rendora/rendora
