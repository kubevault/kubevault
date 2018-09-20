#!/bin/bash

set -x

GOPATH=$(go env GOPATH)
PACKAGE_NAME=github.com/kubevault/operator
REPO_ROOT="$GOPATH/src/$PACKAGE_NAME"
DOCKER_REPO_ROOT="/go/src/$PACKAGE_NAME"
DOCKER_CODEGEN_PKG="/go/src/k8s.io/code-generator"

pushd $REPO_ROOT

rm -rf "$REPO_ROOT"/apis/kubevault/v1alpha1/*.generated.go

docker run --rm -ti -u $(id -u):$(id -g) \
  -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
  -w "$DOCKER_REPO_ROOT" \
  appscode/gengo:release-1.11 "$DOCKER_CODEGEN_PKG"/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/kubevault/operator/client \
  github.com/kubevault/operator/apis \
  kubevault:v1alpha1 \
  --go-header-file "$DOCKER_REPO_ROOT/hack/gengo/boilerplate.go.txt"

# Generate openapi
docker run --rm -ti -u $(id -u):$(id -g) \
  -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
  -w "$DOCKER_REPO_ROOT" \
  appscode/gengo:release-1.11 openapi-gen \
  --v 1 --logtostderr \
  --go-header-file "hack/gengo/boilerplate.go.txt" \
  --input-dirs "$PACKAGE_NAME/apis/kubevault/v1alpha1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/util/intstr,k8s.io/apimachinery/pkg/version,k8s.io/api/core/v1,k8s.io/api/apps/v1,kmodules.xyz/offshoot-api/api/v1,github.com/appscode/go/encoding/json/types" \
  --output-package "$PACKAGE_NAME/apis/kubevault/v1alpha1"

# Generate crds.yaml and swagger.json
go run ./hack/gencrd/main.go

popd
