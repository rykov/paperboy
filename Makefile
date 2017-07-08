VERSION="0.0.0"
#VERSION=$(patsubst v%,%,$(shell git describe --tags))
LDFLAGS=-ldflags "-X 'main.version=$(VERSION) ($(shell date -u +%Y-%m-%d\ %H:%M:%S))'"
default: build

#describe:
#	go run $(LDFLAGS) main.go version

build:
	go build $(LDFLAGS) -v .

install:
	go install $(LDFLAGS) -v .

test:
	go test -race $(LDFLAGS) $(shell go list ./... | grep -v /vendor/)

melody-install: install-melody
	melody install

install-melody:
	go get -u github.com/mdy/melody
	go install github.com/mdy/melody
