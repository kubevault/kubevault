#!/bin/bash

set -x
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/kubevault.dev/operator"

pushd $REPO_ROOT

export APPSCODE_ENV=prod
rm -rf dist
./hack/docker/make.sh
./hack/docker/make.sh release
rm -rf dist/.tag

popd
