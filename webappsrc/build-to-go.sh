#!/bin/bash

echo "Installing go-bindata"
go get -u github.com/go-bindata/go-bindata/...

echo "Building React App"
yarn build

echo "Bundling to ../web"
go-bindata -pkg web -o "../web/webdata.go" -prefix build/ build/...

echo "Done!"

