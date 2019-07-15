#!/bin/bash

set -x

GOPATH=$(go env GOPATH)
PACKAGE_NAME=kubevault.dev/operator
REPO_ROOT="$GOPATH/src/$PACKAGE_NAME"
DOCKER_REPO_ROOT="/go/src/$PACKAGE_NAME"
DOCKER_CODEGEN_PKG="/go/src/k8s.io/code-generator"
apiGroups=(kubevault/v1alpha1 catalog/v1alpha1 policy/v1alpha1 engine/v1alpha1)

pushd $REPO_ROOT

rm -rf "$REPO_ROOT"/apis/kubevault/v1alpha1/*.generated.go
rm -rf "$REPO_ROOT"/apis/catalog/v1alpha1/*.generated.go
rm -rf "$REPO_ROOT"/apis/policy/v1alpha1/*.generated.go
mkdir -p "$REPO_ROOT"/api/api-rules

docker run --rm -ti -u $(id -u):$(id -g) \
  -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
  -w "$DOCKER_REPO_ROOT" \
  appscode/gengo:release-1.14 "$DOCKER_CODEGEN_PKG"/generate-groups.sh all \
  kubevault.dev/operator/client \
  kubevault.dev/operator/apis \
  "kubevault:v1alpha1 catalog:v1alpha1 config:v1alpha1 policy:v1alpha1 engine:v1alpha1" \
  --go-header-file "$DOCKER_REPO_ROOT/hack/gengo/boilerplate.go.txt"

# Generate openapi
for gv in "${apiGroups[@]}"; do
  docker run --rm -ti -u $(id -u):$(id -g) \
    -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
    -w "$DOCKER_REPO_ROOT" \
    appscode/gengo:release-1.14 openapi-gen \
    --v 1 --logtostderr \
    --go-header-file "hack/gengo/boilerplate.go.txt" \
    --input-dirs "$PACKAGE_NAME/apis/${gv},k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/util/intstr,k8s.io/apimachinery/pkg/version,k8s.io/api/core/v1,k8s.io/api/apps/v1,kmodules.xyz/offshoot-api/api/v1,github.com/appscode/go/encoding/json/types,kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1,kmodules.xyz/monitoring-agent-api/api/v1,k8s.io/api/rbac/v1" \
    --output-package "$PACKAGE_NAME/apis/${gv}" \
    --report-filename api/api-rules/violation_exceptions.list
done

# Generate crds.yaml and swagger.json
go run ./hack/gencrd/main.go

popd
