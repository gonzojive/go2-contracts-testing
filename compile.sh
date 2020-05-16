#!/bin/bash

set -e

export PATH=/home/red/git/demos/go2/bin:$PATH

go mod tidy

go tool go2go translate forms2/*.go2
# Fix a bug with how GenericForm gets translated into sexpressions.GenericForm:
sed -i 's/sexpressions\.//g' forms2/*.go
go test forms2/*.go
