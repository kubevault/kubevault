#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubevault/operator/hack/gendocs
go run main.go
popd
