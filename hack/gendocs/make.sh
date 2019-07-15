#!/usr/bin/env bash

pushd $GOPATH/src/kubevault.dev/operator/hack/gendocs
go run main.go
popd
