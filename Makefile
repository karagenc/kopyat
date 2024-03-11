PREFIX ?= /usr/local/bin

all: kopyaship kopyashipd

kopyaship:
	go build -ldflags '-s -w' ./cmd/kopyaship

kopyashipd:
	go build -ldflags '-s -w' ./cmd/kopyashipd

install:
	mv kopyaship kopyashipd $(PREFIX)

.PHONY: help install kopyaship kopyashipd
