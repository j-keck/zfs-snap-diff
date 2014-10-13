#!/bin/sh
#
# script generate 'bindata.go' 
#

echo "generate 'bindata.go'"
go-bindata \
 -ignore=.git \
 -ignore=config.json \
 -ignore=README \
 -ignore=angular-mocks.js \
 -ignore='emacs.*core' \
 webapp/...
