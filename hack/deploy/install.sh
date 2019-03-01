#!/bin/bash
set -eou pipefail

crds=(
  vaultservers.kubevault.com
  vaultserverversions.catalog.kubevault.com
  vaultpolicies.policy.kubevault.com
  vaultpolicybindings.policy.kubevault.com
  databaseaccessrequests.authorization.kubedb.com
  mongodbroles.authorization.kubedb.com
  mysqlroles.authorization.kubedb.com
  postgresroles.authorization.kubedb.com
  awsroles.engine.kubevault.com
  awsaccesskeyrequests.engine.kubevault.com
)
apiServices=(
  v1alpha1.validators.kubevault.com
  v1alpha1.mutators.kubevault.com
  v1alpha1.validators.authorization.kubedb.com
  v1alpha1.validators.engine.kubevault.com
)

echo "checking kubeconfig context"
kubectl config current-context || {
  echo "Set a context (kubectl use-context <context>) out of the following:"
  echo
  kubectl config get-contexts
  exit 1
}
echo ""

# http://redsymbol.net/articles/bash-exit-traps/
function cleanup() {
  rm -rf $ONESSL ca.crt ca.key server.crt server.key
}

export APPSCODE_ENV=${APPSCODE_ENV:-prod}
trap cleanup EXIT

# ref: https://github.com/appscodelabs/libbuild/blob/master/common/lib.sh#L55
inside_git_repo() {
  git rev-parse --is-inside-work-tree >/dev/null 2>&1
  inside_git=$?
  if [ "$inside_git" -ne 0 ]; then
    echo "Not inside a git repository"
    exit 1
  fi
}

detect_tag() {
  inside_git_repo

  # http://stackoverflow.com/a/1404862/3476121
  git_tag=$(git describe --exact-match --abbrev=0 2>/dev/null || echo '')

  commit_hash=$(git rev-parse --verify HEAD)
  git_branch=$(git rev-parse --abbrev-ref HEAD)
  commit_timestamp=$(git show -s --format=%ct)

  if [ "$git_tag" != '' ]; then
    TAG=$git_tag
    TAG_STRATEGY='git_tag'
  elif [ "$git_branch" != 'master' ] && [ "$git_branch" != 'HEAD' ] && [[ "$git_branch" != release-* ]]; then
    TAG=$git_branch
    TAG_STRATEGY='git_branch'
  else
    hash_ver=$(git describe --tags --always --dirty)
    TAG="${hash_ver}"
    TAG_STRATEGY='commit_hash'
  fi

  export TAG
  export TAG_STRATEGY
  export git_tag
  export git_branch
  export commit_hash
  export commit_timestamp
}

onessl_found() {
  # https://stackoverflow.com/a/677212/244009
  if [ -x "$(command -v onessl)" ]; then
    onessl wait-until-has -h >/dev/null 2>&1 || {
      # old version of onessl found
      echo "Found outdated onessl"
      return 1
    }
    export ONESSL=onessl
    return 0
  fi
  return 1
}

onessl_found || {
  echo "Downloading onessl ..."
  # ref: https://stackoverflow.com/a/27776822/244009
  case "$(uname -s)" in
    Darwin)
      curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.10.0/onessl-darwin-amd64
      chmod +x onessl
      export ONESSL=./onessl
      ;;

    Linux)
      curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.10.0/onessl-linux-amd64
      chmod +x onessl
      export ONESSL=./onessl
      ;;

    CYGWIN* | MINGW* | MSYS*)
      curl -fsSL -o onessl.exe https://github.com/kubepack/onessl/releases/download/0.10.0/onessl-windows-amd64.exe
      chmod +x onessl.exe
      export ONESSL=./onessl.exe
      ;;
    *)
      echo 'other OS'
      ;;
  esac
}

# ref: https://stackoverflow.com/a/7069755/244009
# ref: https://jonalmeida.com/posts/2013/05/26/different-ways-to-implement-flags-in-bash/
# ref: http://tldp.org/LDP/abs/html/comparison-ops.html

export VAULT_OPERATOR_NAMESPACE=kube-system
export VAULT_OPERATOR_SERVICE_ACCOUNT=vault-operator
export VAULT_OPERATOR_RUN_ON_MASTER=0
export VAULT_OPERATOR_ENABLE_VALIDATING_WEBHOOK=false
export VAULT_OPERATOR_ENABLE_MUTATING_WEBHOOK=false
export VAULT_OPERATOR_CATALOG=${VAULT_OPERATOR_CATALOG:-all}
export VAULT_OPERATOR_DOCKER_REGISTRY=kubevault
export VAULT_OPERATOR_IMAGE_TAG=0.2.0
export VAULT_OPERATOR_IMAGE_PULL_SECRET=
export VAULT_OPERATOR_IMAGE_PULL_POLICY=IfNotPresent
export VAULT_OPERATOR_ENABLE_ANALYTICS=true
export VAULT_OPERATOR_UNINSTALL=0
export VAULT_OPERATOR_PURGE=0
export VAULT_OPERATOR_ENABLE_STATUS_SUBRESOURCE=false
export VAULT_OPERATOR_BYPASS_VALIDATING_WEBHOOK_XRAY=false
export VAULT_OPERATOR_USE_KUBEAPISERVER_FQDN_FOR_AKS=true
export VAULT_OPERATOR_CLUSTER_NAME=
export VAULT_OPERATOR_PRIORITY_CLASS=system-cluster-critical

export APPSCODE_ENV=${APPSCODE_ENV:-prod}
export SCRIPT_LOCATION="curl -fsSL https://raw.githubusercontent.com/kubevault/operator/0.2.0/"
if [ "$APPSCODE_ENV" = "dev" ]; then
  detect_tag
  export SCRIPT_LOCATION="cat "
  export VAULT_OPERATOR_IMAGE_TAG=$TAG
  export VAULT_OPERATOR_IMAGE_PULL_POLICY=Always
fi

KUBE_APISERVER_VERSION=$(kubectl version -o=json | $ONESSL jsonpath '{.serverVersion.gitVersion}')
$ONESSL semver --check='<1.9.0' $KUBE_APISERVER_VERSION || {
  export VAULT_OPERATOR_ENABLE_VALIDATING_WEBHOOK=true
  export VAULT_OPERATOR_ENABLE_MUTATING_WEBHOOK=true
}
$ONESSL semver --check='<1.11.0' $KUBE_APISERVER_VERSION || { export VAULT_OPERATOR_ENABLE_STATUS_SUBRESOURCE=true; }

export VAULT_OPERATOR_WEBHOOK_SIDE_EFFECTS=
$ONESSL semver --check='<1.12.0' $KUBE_APISERVER_VERSION || { export VAULT_OPERATOR_WEBHOOK_SIDE_EFFECTS='sideEffects: None'; }

MONITORING_AGENT_NONE="none"
MONITORING_AGENT_BUILTIN="prometheus.io/builtin"
MONITORING_AGENT_COREOS_OPERATOR="prometheus.io/coreos-operator"

export MONITORING_AGENT=${MONITORING_AGENT:-$MONITORING_AGENT_NONE}
export MONITOR_OPERATOR=${MONITOR_OPERATOR:-false}
export SERVICE_MONITOR_LABEL_KEY="app"
export SERVICE_MONITOR_LABEL_VALUE="vault-operator"

show_help() {
  echo "install.sh - install Vault operator"
  echo " "
  echo "install.sh [options]"
  echo " "
  echo "options:"
  echo "-h, --help                             show brief help"
  echo "-n, --namespace=NAMESPACE              specify namespace (default: kube-system)"
  echo "    --docker-registry                  docker registry used to pull Vault images (default: kubevault)"
  echo "    --image-pull-secret                name of secret used to pull Vault images"
  echo "    --run-on-master                    run Vault operator on master"
  echo "    --enable-mutating-webhook          enable/disable mutating webhooks for Kubernetes workloads"
  echo "    --enable-validating-webhook        enable/disable validating webhooks for Vault CRDs"
  echo "    --bypass-validating-webhook-xray   if true, bypasses validating webhook xray checks"
  echo "    --enable-status-subresource        if enabled, uses status sub resource for crds"
  echo "    --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)"
  echo "    --enable-analytics                 send usage events to Google Analytics (default: true)"
  echo "    --uninstall                        uninstall Vault operator"
  echo "    --purge                            purges Vault CRD objects and crds"
  echo "    --install-catalog                  installs Vault server version catalog (default: all)"
  echo "    --monitoring-agent                 specify which monitoring agent to use (default: none)"
  echo "    --monitor-operator                 specify whether to monitor Vault operator (default: false)"
  echo "    --prometheus-namespace             specify the namespace where Prometheus server is running or will be deployed (default: same namespace as vault-operator)"
  echo "    --servicemonitor-label             specify the label for ServiceMonitor crd. Prometheus crd will use this label to select the ServiceMonitor. (default: 'app: vault-operator')"
  echo "    --cluster-name                     Name of cluster used in a multi-cluster setup"
}

while test $# -gt 0; do
  case "$1" in
    -h | --help)
      show_help
      exit 0
      ;;
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
      export VAULT_OPERATOR_NAMESPACE=$(echo $1 | sed -e 's/^[^=]*=//g')
      shift
      ;;
    --docker-registry*)
      export VAULT_OPERATOR_DOCKER_REGISTRY=$(echo $1 | sed -e 's/^[^=]*=//g')
      shift
      ;;
    --image-pull-secret*)
      secret=$(echo $1 | sed -e 's/^[^=]*=//g')
      export VAULT_OPERATOR_IMAGE_PULL_SECRET="name: '$secret'"
      shift
      ;;
    --enable-mutating-webhook*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export VAULT_OPERATOR_ENABLE_MUTATING_WEBHOOK=false
      fi
      shift
      ;;
    --enable-validating-webhook*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export VAULT_OPERATOR_ENABLE_VALIDATING_WEBHOOK=false
      fi
      shift
      ;;
    --bypass-validating-webhook-xray*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export VAULT_OPERATOR_BYPASS_VALIDATING_WEBHOOK_XRAY=false
      else
        export VAULT_OPERATOR_BYPASS_VALIDATING_WEBHOOK_XRAY=true
      fi
      shift
      ;;
    --enable-status-subresource*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export VAULT_OPERATOR_ENABLE_STATUS_SUBRESOURCE=false
      fi
      shift
      ;;
    --use-kubeapiserver-fqdn-for-aks*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export VAULT_OPERATOR_USE_KUBEAPISERVER_FQDN_FOR_AKS=false
      else
        export VAULT_OPERATOR_USE_KUBEAPISERVER_FQDN_FOR_AKS=true
      fi
      shift
      ;;
    --enable-analytics*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export VAULT_OPERATOR_ENABLE_ANALYTICS=false
      fi
      shift
      ;;
    --install-catalog*)
      shift
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export VAULT_OPERATOR_CATALOG=false
      fi
      ;;
    --run-on-master)
      export VAULT_OPERATOR_RUN_ON_MASTER=1
      shift
      ;;
    --uninstall)
      export VAULT_OPERATOR_UNINSTALL=1
      shift
      ;;
    --purge)
      export VAULT_OPERATOR_PURGE=1
      shift
      ;;
    --monitoring-agent*)
       val=$(echo $1 | sed -e 's/^[^=]*=//g')
       if [ "$val" != "$MONITORING_AGENT_BUILTIN" ] && [ "$val" != "$MONITORING_AGENT_COREOS_OPERATOR" ]; then
         echo 'Invalid monitoring agent. Use "builtin" or "coreos-operator"'
         exit 1
       else
         export MONITORING_AGENT="$val"
       fi
       shift
       ;;
     --monitor-operator*)
       val=$(echo $1 | sed -e 's/^[^=]*=//g')
       if [ "$val" = "true" ]; then
         export MONITOR_OPERATOR="$val"
       fi
       shift
       ;;
     --prometheus-namespace*)
       export PROMETHEUS_NAMESPACE=$(echo $1 | sed -e 's/^[^=]*=//g')
       shift
       ;;
     --servicemonitor-label*)
       label=$(echo $1 | sed -e 's/^[^=]*=//g')
       # split label into key value pair
       IFS='='
       pair=($label)
       unset IFS
       # check if the label is valid
       if [ ! ${#pair[@]} = 2 ]; then
         echo "Invalid ServiceMonitor label format. Use '--servicemonitor-label=key=value'"
         exit 1
       fi
       export SERVICE_MONITOR_LABEL_KEY="${pair[0]}"
       export SERVICE_MONITOR_LABEL_VALUE="${pair[1]}"
       shift
       ;;
     --cluster-name*)
       export VAULT_OPERATOR_CLUSTER_NAME=$(echo $1 | sed -e 's/^[^=]*=//g')
       shift
       ;;
    *)
      echo "Error: unknown flag:" $1
      show_help
      exit 1
      ;;
  esac
done

export PROMETHEUS_NAMESPACE=${PROMETHEUS_NAMESPACE:-$VAULT_OPERATOR_NAMESPACE}

if [ "$VAULT_OPERATOR_NAMESPACE" != "kube-system" ]; then
    export VAULT_OPERATOR_PRIORITY_CLASS=""
fi

if [ "$VAULT_OPERATOR_UNINSTALL" -eq 1 ]; then
  # delete webhooks and apiservices
  kubectl delete validatingwebhookconfiguration -l app=vault-operator || true
  kubectl delete mutatingwebhookconfiguration -l app=vault-operator || true
  kubectl delete apiservice -l app=vault-operator
  # delete Vault operator
  kubectl delete deployment -l app=vault-operator --namespace $VAULT_OPERATOR_NAMESPACE
  kubectl delete service -l app=vault-operator --namespace $VAULT_OPERATOR_NAMESPACE
  kubectl delete secret -l app=vault-operator --namespace $VAULT_OPERATOR_NAMESPACE
  # delete RBAC objects, if --rbac flag was used.
  kubectl delete serviceaccount -l app=vault-operator --namespace $VAULT_OPERATOR_NAMESPACE
  kubectl delete clusterrolebindings -l app=vault-operator
  kubectl delete clusterrole -l app=vault-operator
  kubectl delete rolebindings -l app=vault-operator --namespace $VAULT_OPERATOR_NAMESPACE
  kubectl delete role -l app=vault-operator --namespace $VAULT_OPERATOR_NAMESPACE

  # delete servicemonitor and vault-operator-apiserver-cert secret. ignore error as they might not exist
  kubectl delete servicemonitor vault-operator-servicemonitor --namespace $PROMETHEUS_NAMESPACE || true
  kubectl delete secret vault-operator-apiserver-cert --namespace $PROMETHEUS_NAMESPACE || true

  echo "waiting for Vault operator pod to stop running"
  for (( ; ; )); do
    pods=($(kubectl get pods --namespace $VAULT_OPERATOR_NAMESPACE -l app=vault-operator -o jsonpath='{range .items[*]}{.metadata.name} {end}'))
    total=${#pods[*]}
    if [ $total -eq 0 ]; then
      break
    fi
    sleep 2
  done

  # https://github.com/kubernetes/kubernetes/issues/60538
  if [ "$VAULT_OPERATOR_PURGE" -eq 1 ]; then
    for crd in "${crds[@]}"; do
      pairs=($(kubectl get ${crd} --all-namespaces -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.namespace} {end}' || true))
      total=${#pairs[*]}

      # save objects
      if [ $total -gt 0 ]; then
        echo "dumping ${crd} objects into ${crd}.yaml"
        kubectl get ${crd} --all-namespaces -o yaml >${crd}.yaml
      fi

      for ((i = 0; i < $total; i++)); do
        name=${pairs[$i]}
        namespace="default"
        if [[ $crd != *"catalog.kubevault.com" ]]; then
          namespace=${pairs[$i + 1]}
          i+=1
        fi
        # remove finalizers
        kubectl patch ${crd} $name -n $namespace -p '{"metadata":{"finalizers":[]}}' --type=merge || true
        # delete crd object
        echo "deleting ${crd} $namespace/$name"
        kubectl delete ${crd} $name -n $namespace --ignore-not-found=true
      done

      # delete crd
      kubectl delete crd ${crd} --ignore-not-found=true
    done

    # delete user roles
    kubectl delete clusterroles kubevault:core:edit kubevault:core:view --ignore-not-found=true
  fi

  echo
  echo "Successfully uninstalled Vault operator!"
  exit 0
fi

echo "checking whether extended apiserver feature is enabled"
$ONESSL has-keys configmap --namespace=kube-system --keys=requestheader-client-ca-file extension-apiserver-authentication || {
  echo "Set --requestheader-client-ca-file flag on Kubernetes apiserver"
  exit 1
}
echo ""

export KUBE_CA=
export VAULT_OPERATOR_ENABLE_APISERVER=false
if [ "$VAULT_OPERATOR_ENABLE_VALIDATING_WEBHOOK" = true ] || [ "$VAULT_OPERATOR_ENABLE_MUTATING_WEBHOOK" = true ]; then
  $ONESSL get kube-ca >/dev/null 2>&1 || {
    echo "Admission webhooks can't be used when kube apiserver is accesible without verifying its TLS certificate (insecure-skip-tls-verify : true)."
    echo
    exit 1
  }
  export KUBE_CA=$($ONESSL get kube-ca | $ONESSL base64)
  export VAULT_OPERATOR_ENABLE_APISERVER=true
fi

env | sort | grep VAULT_OPERATOR*
echo ""

# create necessary TLS certificates:
# - a local CA key and cert
# - a webhook server key and cert signed by the local CA
$ONESSL create ca-cert
$ONESSL create server-cert server --domains=vault-operator.$VAULT_OPERATOR_NAMESPACE.svc
export SERVICE_SERVING_CERT_CA=$(cat ca.crt | $ONESSL base64)
export TLS_SERVING_CERT=$(cat server.crt | $ONESSL base64)
export TLS_SERVING_KEY=$(cat server.key | $ONESSL base64)

${SCRIPT_LOCATION}hack/deploy/operator.yaml | $ONESSL envsubst | kubectl apply -f -

${SCRIPT_LOCATION}hack/deploy/service-account.yaml | $ONESSL envsubst | kubectl apply -f -
${SCRIPT_LOCATION}hack/deploy/rbac-list.yaml | $ONESSL envsubst | kubectl auth reconcile -f -
${SCRIPT_LOCATION}hack/deploy/user-roles.yaml | $ONESSL envsubst | kubectl auth reconcile -f -
${SCRIPT_LOCATION}hack/deploy/appcatalog-user-roles.yaml | $ONESSL envsubst | kubectl auth reconcile -f -

if [ "$VAULT_OPERATOR_RUN_ON_MASTER" -eq 1 ]; then
  kubectl patch deploy vault-operator -n $VAULT_OPERATOR_NAMESPACE \
    --patch="$(${SCRIPT_LOCATION}hack/deploy/run-on-master.yaml)"
fi

if [ "$VAULT_OPERATOR_ENABLE_APISERVER" = true ]; then
  ${SCRIPT_LOCATION}hack/deploy/apiservices.yaml | $ONESSL envsubst | kubectl apply -f -
fi
if [ "$VAULT_OPERATOR_ENABLE_VALIDATING_WEBHOOK" = true ]; then
  ${SCRIPT_LOCATION}hack/deploy/validating-webhook.yaml | $ONESSL envsubst | kubectl apply -f -
fi
if [ "$VAULT_OPERATOR_ENABLE_MUTATING_WEBHOOK" = true ]; then
  ${SCRIPT_LOCATION}hack/deploy/mutating-webhook.yaml | $ONESSL envsubst | kubectl apply -f -
fi

echo
echo "waiting until Vault operator deployment is ready"
$ONESSL wait-until-ready deployment vault-operator --namespace $VAULT_OPERATOR_NAMESPACE || {
  echo "Vault operator deployment failed to be ready"
  exit 1
}

if [ "$VAULT_OPERATOR_ENABLE_APISERVER" = true ]; then
  echo "waiting until Vault operator apiservice is available"
  for api in "${apiServices[@]}"; do
    $ONESSL wait-until-ready apiservice ${api} || {
      echo "Vault operator apiservice $api failed to be ready"
      exit 1
    }
  done
fi

echo "waiting until Vault operator crds are ready"
for crd in "${crds[@]}"; do
  $ONESSL wait-until-ready crd ${crd} || {
    echo "$crd crd failed to be ready"
    exit 1
  }
done

if [ "$VAULT_OPERATOR_CATALOG" = "all" ] || [ "$VAULT_OPERATOR_CATALOG" = "vaultserver" ]; then
  echo "installing Vault server catalog"
  ${SCRIPT_LOCATION}hack/deploy/catalog/vaultserver.yaml | $ONESSL envsubst | kubectl apply -f -
  echo
fi

if [ "$VAULT_OPERATOR_ENABLE_VALIDATING_WEBHOOK" = true ]; then
  echo "checking whether admission webhook(s) are activated or not"
  active=$($ONESSL wait-until-has annotation \
    --apiVersion=apiregistration.k8s.io/v1beta1 \
    --kind=APIService \
    --name=v1alpha1.validators.kubevault.com \
    --key=admission-webhook.appscode.com/active \
    --timeout=5m || {
    echo
    echo "Failed to check if admission webhook(s) are activated or not. Please check operator logs to debug further."
    exit 1
  })
  if [ "$active" = false ]; then
    echo
    echo "Admission webhooks are not activated."
    echo "Enable it by configuring --enable-admission-plugins flag of kube-apiserver."
    echo "For details, visit: https://appsco.de/kube-apiserver-webhooks ."
    echo "After admission webhooks are activated, please uninstall and then reinstall Vault operator."
    # uninstall misconfigured webhooks to avoid failures
    kubectl delete validatingwebhookconfiguration -l app=vault-operator || true
    exit 1
  fi
fi

# configure prometheus monitoring
if [ "$MONITOR_OPERATOR" = "true" ] && [ "$MONITORING_AGENT" != "$MONITORING_AGENT_NONE" ]; then
  # if operator monitoring is enabled and prometheus-namespace is provided,
  # create vault-operator-apiserver-cert there. this will be mounted on prometheus pod.
  if [ "$PROMETHEUS_NAMESPACE" != "$VAULT_OPERATOR_NAMESPACE" ]; then
    ${SCRIPT_LOCATION}hack/deploy/monitor/apiserver-cert.yaml | $ONESSL envsubst | kubectl apply -f -
  fi

  case "$MONITORING_AGENT" in
    "$MONITORING_AGENT_BUILTIN")
      kubectl annotate service vault-operator -n "$VAULT_OPERATOR_NAMESPACE" --overwrite \
        prometheus.io/scrape="true" \
        prometheus.io/path="/metrics" \
        prometheus.io/port="8443" \
        prometheus.io/scheme="https"
      ;;
    "$MONITORING_AGENT_COREOS_OPERATOR")
      ${SCRIPT_LOCATION}hack/deploy/monitor/servicemonitor.yaml | $ONESSL envsubst | kubectl apply -f -
      ;;
  esac
fi

echo
echo "Successfully installed Vault operator in $VAULT_OPERATOR_NAMESPACE namespace!"
