##=======================================================================##
## Makefile
## Created: Wed Aug 05 14:35:14 PDT 2015 @941 /Internet Time/
# :mode=makefile:tabSize=3:indentSize=3:
## Purpose:
##======================================================================##

SHELL=/bin/bash
PROJECT_NAME = DiffStore
GPATH = $(shell pwd)

.PHONY: fmt get-deps test install build scrape clean

install: fmt get-deps test
	@GOPATH=${GPATH} go build *.go

build: fmt get-deps test
	@GOPATH=${GPATH} go build *.go

get-deps:
	mkdir -p "src"
	mkdir -p "pkg"
	@GOPATH=${GPATH} go get github.com/boltdb/bolt
	@GOPATH=${GPATH} go get github.com/paulmach/go.geojson
	@GOPATH=${GPATH} go get github.com/sergi/go-diff/diffmatchpatch
	@GOPATH=${GPATH} go get github.com/sjsafranek/SkeletonDB
	@GOPATH=${GPATH} go get github.com/sjsafranek/DiffStore

fmt:
	@GOPATH=${GPATH} gofmt -s -w *.go

test: fmt get-deps
	@GOPATH=${GPATH} go test -v -bench=. -test.benchmem

scrape:
	@find src -type d -name '.hg' -or -type d -name '.git' | xargs rm -rf

clean:
	@GOPATH=${GPATH} go clean
