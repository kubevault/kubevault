#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kube-vault/operator/hack/gendocs
go run main.go
popd
