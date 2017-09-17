#!/usr/bin/env bash

pushd $GOPATH/src/github.com/appscode/steward/hack/gendocs
go run main.go
popd
