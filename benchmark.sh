#!/bin/bash

export GOPATH="`pwd`"

cd skeleton_db
go test -bench=. -test.benchmem
