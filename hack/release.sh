#!/bin/bash

set -x
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/appscode/steward"
rm -rf $REPO_ROOT/dist

APPSCODE_ENV=prod ./hack/docker/setup.sh
APPSCODE_ENV=prod ./hack/docker/setup.sh release

rm $REPO_ROOT/dist/.tag
