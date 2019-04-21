#!/bin/bash

TAG=`git describe --exact-match --tags HEAD`

if [[ $? -eq 0 ]];
then
  echo "Releasing for tag ${TAG}"
  echo "Downloading deps"
  go get -v
  echo "Dowloading gox for multi-arch"
  go get github.com/mitchellh/gox
  mkdir -p out
  mkdir -p bins
  echo "Multi-arch build"
  gox -output "out/{{.OS}}-{{.Arch}}/{{.Dir}}" -arch="arm arm64 386 amd64" -os="windows linux"
  gox -output "out/{{.OS}}-{{.Arch}}/{{.Dir}}" -arch="amd64" -os="darwin"
  cd out
  for i in *
  do
    cp ../qo100.sample.toml $i/
    echo "Zipping qo100-dedrift-${i}.zip"
    zip -r ../bins/qo100-dedrift-$i.zip $i/*
  done
  cd ..
  ls -la bins
else
  echo "No tags for current commit. Skipping releases."
fi
