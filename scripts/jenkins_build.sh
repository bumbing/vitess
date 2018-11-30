#!/bin/bash

set -x
set -e

export VT_ARTIFACTS=${VT_ARTIFACTS:-'vitess,vtgate,vtworker'}
export BUILD_DIR=$WORKSPACE/BUILD_DIR

if [[ -d $BUILD_DIR ]]; then 
    rm -rf $BUILD_DIR/*
else
    mkdir $BUILD_DIR
fi

source $HOME/.aws/iam_keys

# Create a build artifact and docker image based on the //:full_dist bazel target.
cd $WORKSPACE

export PACKAGE_DEB=$PACKAGE_DEB
export SKIP_BUILD=false

set +e
aws ecr create-repository --repository-name vitess --region us-east-1
aws ecr create-repository --repository-name vitess/base --region us-east-1
set -e

export TARBALL_FN_GZ=vitess-${GIT_COMMIT:0:7}.tar.gz
export TARBALL_GZ="${BUILD_DIR}/${TARBALL_FN_GZ}"
# Do all the work of building
./scripts/build_pinterest_docker.sh

export BUILT_ARTIFACTS=''
# Package telefig for each vitess artifact
IFS=',';for ARTIFACT in ${VT_ARTIFACTS}; do
export TARBALL_FN_GZ=${ARTIFACT}-${GIT_COMMIT:0:7}.tar.gz
export TARBALL_GZ="${BUILD_DIR}/${TARBALL_FN_GZ}"
export TARBALL_SRC="${WORKSPACE}/teletraan/${ARTIFACT}"
if [[ -d ${WORKSPACE}/teletraan/config && $ARTIFACT -eq 'vitess' ]]; then
    # TODO remove `vitess` artifact once the PR is merged
    export TARBALL_SRC="${WORKSPACE}/teletraan"
fi
if [[ -d $TARBALL_SRC ]]; then
    echo "packaging artifact:${ARTIFACT} to:${TARBALL_GZ} using telefig from:${TARBALL_SRC}"
    tar czvf $TARBALL_GZ -C $TARBALL_SRC .
    s3up $TARBALL_GZ pinterest-builds vitess/$TARBALL_FN_GZ
    export BUILT_ARTIFACTS="${BUILT_ARTIFACTS},${ARTIFACT}"
else
    echo "bypassing artifact:${ARTIFACT} due to missing telefig at:${TARBALL_SRC}"
fi
done

if [[ ${#BUILT_ARTIFACTS} -eq 0 ]]; then
    echo "there's no artifact built from: ${VT_ARTIFACTS}, check the telefig structure!"
    exit 1
else
    # stipping off head comma
    export BUILT_ARTIFACTS="${BUILT_ARTIFACTS:1:${#BUILT_ARTIFACTS}-1}"
    echo "bundle the following artifacts: ${BUILT_ARTIFACTS}"
fi

# end
cat << EOF > "$WORKSPACE/builds.property"
BUILDS_NAME=${BUILT_ARTIFACTS}
COMMIT=$GIT_COMMIT
REPO=VT
TYPE=pinterest-builds/vitess
BRANCH=${GIT_BRANCH#origin/}
TARGET_DIR=vitess
BUILD_URL=$BUILD_URL
COMMIT_DATE=$(($(git show -s --format=%ct ${COMMIT}) * 1000))
EOF
