PREFIX ?= /usr/local/bin

all: kopyaship kopyashipd

kopyaship:
	go build -ldflags '-s -w' ./cmd/kopyaship

kopyashipd:
	go build -ldflags '-s -w' ./cmd/kopyashipd

test:
	go test -v -buildmode=default -race ./...

test-coverage:
	go test -buildmode=default -coverprofile coverage.out -covermode=atomic ./...

install:
	mv kopyaship kopyashipd $(PREFIX)

clean:
	rm -f kopyaship kopyashipd *.exe ifile/test_ifile_* ifile/test_txtfile_*

.PHONY: kopyaship kopyashipd test test-coverage install clean
