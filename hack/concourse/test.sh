#!/usr/bin/env bash

set -eoux pipefail

ORG_NAME=kubevault
REPO_NAME=operator
APP_LABEL=vault #required for `kubectl describe deploy -n kube-system -l app=$APP_LABEL`

export DOCKER_REGISTRY=appscodeci

# get concourse-common
pushd $REPO_NAME
git status # required, otherwise you'll get error `Working tree has modifications.  Cannot add.`. why?
git subtree pull --prefix hack/libbuild https://github.com/appscodelabs/libbuild.git master --squash -m 'concourse'
popd

source $REPO_NAME/hack/libbuild/concourse/init.sh

cp creds/kubevault/gcs.json /gcs.json
cp creds/kubevault/.env $GOPATH/src/github.com/$ORG_NAME/$REPO_NAME/hack/config/.env

pushd $GOPATH/src/github.com/$ORG_NAME/$REPO_NAME
source hack/config/.env
rm hack/config/.env

# install dependencies
./hack/builddeps.sh

# run tests
ginkgo -v test/e2e

popd
