# Main Makefile for archive
#
# Copyright 2018 © by Ollivier Robert
#

GO=		go
GOBIN=  ${GOPATH}/bin

SRCS= archive.go utils.go

OPTS=	-ldflags="-s -w" -v

all: build

build: ${SRCS}
	${GO} build -o ${BIN} ${OPTS} .

test:
	${GO} test -v .

install: ${BIN}
	${GO} install ${OPTS} .

clean:
	${GO} clean -v

push:
	git push --all
	git push --tags
