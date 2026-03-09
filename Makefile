VERSION ?= dev
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test lint clean install

build:
	go build $(LDFLAGS) -o save .

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -f save

install: build
	cp save /usr/local/bin/
