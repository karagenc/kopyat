PREFIX ?= /usr/local/bin

all: generate build

generate:
	go run ./scripts/generate.go
	statik -f -src=. -include=kopyaship_example.yml -dest=./internal

build:
	go build -ldflags '-s -w' ./cmd/kopyaship

test:
	go test -v -buildmode=default -race ./...

test-coverage:
	go test -buildmode=default -coverprofile coverage.out -covermode=atomic ./...

install:
	mv kopyaship $(PREFIX)

clean:
	rm -f kopyaship kopyaship.exe ifile/test_ifile_* ifile/test_txtfile_*

.PHONY: build generate test test-coverage install clean
