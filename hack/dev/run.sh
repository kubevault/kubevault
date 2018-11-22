#!/bin/bash
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/kubevault/operator"

pushd $REPO_ROOT

# http://redsymbol.net/articles/bash-exit-traps/
function cleanup() {
  rm -rf $ONESSL ca.crt ca.key server.crt server.key
}
trap cleanup EXIT

# https://stackoverflow.com/a/677212/244009
if [[ ! -z "$(command -v onessl)" ]]; then
  export ONESSL=onessl
else
  # ref: https://stackoverflow.com/a/27776822/244009
  case "$(uname -s)" in
    Darwin)
      curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.7.0/onessl-darwin-amd64
      chmod +x onessl
      export ONESSL=./onessl
      ;;

    Linux)
      curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.7.0/onessl-linux-amd64
      chmod +x onessl
      export ONESSL=./onessl
      ;;

    CYGWIN* | MINGW32* | MSYS*)
      curl -fsSL -o onessl.exe https://github.com/kubepack/onessl/releases/download/0.7.0/onessl-windows-amd64.exe
      chmod +x onessl.exe
      export ONESSL=./onessl.exe
      ;;
    *)
      echo 'other OS'
      ;;
  esac
fi

export VAULT_OPERATOR_NAMESPACE=default
export KUBE_CA=$($ONESSL get kube-ca | $ONESSL base64)
export VAULT_OPERATOR_ENABLE_WEBHOOK=true
export VAULT_OPERATOR_E2E_TEST=false
export VAULT_OPERATOR_DOCKER_REGISTRY=kubevault
export VAULT_OPERATOR_ENABLE_SUBRESOURCE=false

while test $# -gt 0; do
  case "$1" in
    -n)
      shift
      if test $# -gt 0; then
        export VAULT_OPERATOR_NAMESPACE=$1
      else
        echo "no namespace specified"
        exit 1
      fi
      shift
      ;;
    --namespace*)
      shift
      if test $# -gt 0; then
        export VAULT_OPERATOR_NAMESPACE=$1
      else
        echo "no namespace specified"
        exit 1
      fi
      shift
      ;;
    --enable-webhook*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export VAULT_OPERATOR_ENABLE_WEBHOOK=false
      fi
      shift
      ;;
    --docker-registry*)
      export VAULT_OPERATOR_DOCKER_REGISTRY=$(echo $1 | sed -e 's/^[^=]*=//g')
      shift
      ;;
    --test*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "true" ]; then
        export VAULT_OPERATOR_E2E_TEST=true
      fi
      shift
      ;;
    --enable-subresource*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "true" ]; then
        export VAULT_OPERATOR_ENABLE_SUBRESOURCE=true
      fi
      shift
      ;;
    *)
      echo $1
      exit 1
      ;;
  esac
done

# !!! WARNING !!! Never do this in prod cluster
kubectl create clusterrolebinding anonymous-cluster-admin --clusterrole=cluster-admin --user=system:anonymous || true

kubectl create -R -f $REPO_ROOT/api/crds || true

cat $REPO_ROOT/hack/deploy/catalog/vaultserver.yaml | $ONESSL envsubst | kubectl apply -f -
cat $REPO_ROOT/hack/dev/apiregistration.yaml | $ONESSL envsubst | kubectl apply -f -
cat $REPO_ROOT/hack/deploy/validating-webhook.yaml | $ONESSL envsubst | kubectl apply -f -
cat $REPO_ROOT/hack/deploy/mutating-webhook.yaml | $ONESSL envsubst | kubectl apply -f -
rm -f ./onessl

$REPO_ROOT/hack/make.py

if [ "$VAULT_OPERATOR_E2E_TEST" = false ]; then # don't run operator while run this script from test
vault-operator run --v=3 \
  --secure-port=8443 \
  --enable-status-subresource="$VAULT_OPERATOR_ENABLE_SUBRESOURCE" \
  --kubeconfig="$HOME/.kube/config" \
  --authorization-kubeconfig="$HOME/.kube/config" \
  --authentication-kubeconfig="$HOME/.kube/config" \
  --authentication-skip-lookup
fi

popd
