#!/bin/sh
#
# this script triggers the webapp build, imports the
# generated code in this repo and commits the change
#
cp -v $(nix-build -A bindata)/bindata.go pkg/webapp/bindata.go
git add pkg/webapp/bindata.go
git commit -m 'regen bindata.go'
