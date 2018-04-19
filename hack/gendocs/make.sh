#!/usr/bin/env bash

pushd $GOPATH/src/github.com/soter/vault-operator/hack/gendocs
go run main.go
popd
