PREFIX ?= /usr/local/bin

all: generate build

generate:
	go run ./scripts/generate.go
	statik -f -src=. -include=kopyat_example.yml -dest=./internal

build:
	go build -ldflags '-s -w' ./cmd/kopyat

test:
	go test -v -buildmode=default -race ./...

test-coverage:
	go test -buildmode=default -coverprofile coverage.out -covermode=atomic ./...

install:
	mv kopyat $(PREFIX)

clean:
	rm -f kopyat kopyat.exe ifile/test_ifile_* ifile/test_txtfile_*

.PHONY: build generate test test-coverage install clean
