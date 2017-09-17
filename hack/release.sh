#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

pushd "$(go env GOPATH)/src/github.com/appscode/steward"
rm -rf dist
APPSCODE_ENV=prod ./hack/docker/setup.sh
APPSCODE_ENV=prod ./hack/docker/setup.sh release
popd
