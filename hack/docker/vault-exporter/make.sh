#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubevault}

GOPATH=$(go env GOPATH)
REPO_ROOT=$GOPATH/src/github.com/kubevault/operator

source "$REPO_ROOT/hack/libbuild/common/lib.sh"
source "$REPO_ROOT/hack/libbuild/common/kubevault_image.sh"

IMG_REGISTRY=prom
IMG=statsd-exporter
TAG=latest

docker pull "$IMG_REGISTRY/$IMG:$TAG"

docker tag "$IMG_REGISTRY/$IMG:$TAG" "$DOCKER_REGISTRY/$IMG:$TAG"

docker push "$DOCKER_REGISTRY/$IMG:$TAG"

DIST="$REPO_ROOT/dist"
mkdir -p "$DIST"

build() {
  pushd "$REPO_ROOT/hack/docker/vault-exporter"

  local cmd="docker build -t $DOCKER_REGISTRY/$IMG:$TAG ."
  echo $cmd; $cmd

  popd
}

source_repo $@
