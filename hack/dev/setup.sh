#!/usr/bin/env bash
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/kubevault/operator"

pushd $REPO_ROOT

# https://stackoverflow.com/a/677212/244009
if  [[ ! -z "$(command -v onessl)" ]]; then
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

        CYGWIN*|MINGW32*|MSYS*)
            curl -fsSL -o onessl.exe https://github.com/kubepack/onessl/releases/download/0.7.0/onessl-windows-amd64.exe
            chmod +x onessl.exe
            export ONESSL=./onessl.exe
            ;;
        *)
            echo 'other OS'
            ;;
    esac
fi

export KUBEVAULT_NAMESPACE=kube-system
export KUBE_CA=$($ONESSL get kube-ca | $ONESSL base64)
while test $# -gt 0; do
    case "$1" in
        -n)
            shift
            if test $# -gt 0; then
                export KUBEVAULT_NAMESPACE=$1
            else
                echo "no namespace specified"
                exit 1
            fi
            shift
            ;;
        --namespace*)
            shift
            if test $# -gt 0; then
                export KUBEVAULT_NAMESPACE=$1
            else
                echo "no namespace specified"
                exit 1
            fi
            shift
            ;;
         *)
            echo $1
            exit 1
            ;;
    esac
done

cat $REPO_ROOT/hack/dev/apiregistration.yaml | $ONESSL envsubst | kubectl apply -f -
cat $REPO_ROOT/hack/deploy/validating-webhook.yaml | $ONESSL envsubst | kubectl apply -f -
# cat $REPO_ROOT/hack/deploy/mutating-webhook.yaml | $ONESSL envsubst | kubectl apply -f -
rm -f ./onessl

$REPO_ROOT/hack/make.py
vault-operator run --v=4 \
  --secure-port=8443 \
  --enable-status-subresource=true \
  --kubeconfig="$HOME/.kube/config" \
  --authorization-kubeconfig="$HOME/.kube/config" \
  --authentication-kubeconfig="$HOME/.kube/config" \
  --authentication-skip-lookup
popd