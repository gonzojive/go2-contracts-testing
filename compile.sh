#!/bin/bash

set -e

export PATH=/home/red/git/demos/go2/bin:$PATH

go mod tidy

go tool go2go translate forms2/*.go2
go test forms2/*.go
