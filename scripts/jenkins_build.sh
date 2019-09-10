#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export VT_ARTIFACTS=${VT_ARTIFACTS:-'vitess,vtgate,vtworker,vtctld'}

# Create a build artifact and docker image based on the //:full_dist bazel target.
cd "${WORKSPACE}"
BUILD_DIR=$(mktemp -d -t "vitess-build-${BUILD_NUMBER:-0}-XXXXXXXXXX" --tmpdir=.)
BUILD_DIR=${BUILD_DIR:2} # that's for stripping `./` prefix
export BUILD_DIR
echo "use temporary directory: ${BUILD_DIR} for this build: ${BUILD_NUMBER:-0}"

source "${HOME}"/.aws/iam_keys

export PACKAGE_DEB=${PACKAGE_DEB:-false}
export SKIP_BUILD=false

set +e
aws ecr create-repository --repository-name vitess --region us-east-1
aws ecr create-repository --repository-name vitess/base --region us-east-1
aws ecr create-repository --repository-name vitess/vtctld --region us-east-1
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
if [[ -d $TARBALL_SRC ]]; then
    echo "packaging artifact:${ARTIFACT} to:${TARBALL_GZ} using telefig from:${TARBALL_SRC}"
    tar -czvf "${TARBALL_GZ}" -C "${TARBALL_SRC}" .
    s3up "${TARBALL_GZ}" pinterest-builds vitess/"${TARBALL_FN_GZ}"
    export BUILT_ARTIFACTS="${BUILT_ARTIFACTS},${ARTIFACT}"
else
    echo "bypassing artifact:${ARTIFACT} due to missing telefig at:${TARBALL_SRC}"
fi
done

if [[ ${#BUILT_ARTIFACTS} -eq 0 ]]; then
    echo "there's no artifact built from: ${VT_ARTIFACTS}, check the telefig structure!"
    exit 1
else
    # stripping off head comma
    export BUILT_ARTIFACTS="${BUILT_ARTIFACTS:1:${#BUILT_ARTIFACTS}-1}"
    echo "bundle the following artifacts: ${BUILT_ARTIFACTS}"
fi

# end
cat << EOF > "${WORKSPACE}/builds.property"
BUILDS_NAME=${BUILT_ARTIFACTS}
COMMIT=$GIT_COMMIT
REPO=VT
TYPE=pinterest-builds/vitess
BRANCH=${GIT_BRANCH#origin/}
TARGET_DIR=vitess
BUILD_URL=$BUILD_URL
COMMIT_DATE=$(($(git show -s --format=%ct "${GIT_COMMIT}") * 1000))
EOF
