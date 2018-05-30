#!/bin/bash

set -x
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/kube-vault/operator"

pushd $REPO_ROOT

export APPSCODE_ENV=prod
rm -rf dist
./hack/docker/make.sh
./hack/docker/make.sh release
rm -rf dist/.tag

popd
