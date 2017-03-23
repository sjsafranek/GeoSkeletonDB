##=======================================================================##
## Makefile
## Created: Wed Aug 05 14:35:14 PDT 2015 @941 /Internet Time/
# :mode=makefile:tabSize=3:indentSize=3:
## Purpose: 
##======================================================================##

SHELL=/bin/bash
PROJECT_NAME = GeoSkeletonDB
GPATH = $(shell pwd)

.PHONY: fmt deps test install build scrape clean

install: fmt deps
	@GOPATH=${GPATH} go build -o db-cli client.go

build: fmt deps
	@GOPATH=${GPATH} go build -o db-cli client.go

deps:
	mkdir -p "src"
	mkdir -p "pkg"
	@GOPATH=${GPATH} go get github.com/paulmach/go.geojson
	@GOPATH=${GPATH} go get github.com/sjsafranek/SkeletonDB

fmt:
	@GOPATH=${GPATH} gofmt -s -w geo_skeleton_db

test:
	./benchmark.sh

scrape:
	@find src -type d -name '.hg' -or -type d -name '.git' | xargs rm -rf

clean:
	@GOPATH=${GPATH} go clean
