#!/bin/sh

set -ex

go get -t -v -d
go test -v
