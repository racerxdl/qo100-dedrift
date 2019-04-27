#!/bin/sh

protoc -I ./ dedrift.proto --go_out=plugins=grpc:.

