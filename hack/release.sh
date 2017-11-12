#!/bin/bash

set -x
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/appscode/steward"

pushd $REPO_ROOT

rm -rf dist
APPSCODE_ENV=prod ./hack/docker/setup.sh
APPSCODE_ENV=prod ./hack/docker/setup.sh release
rm dist/.tag

popd
