#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Relevant env variables:
#
# REBUILD_COMMON: "true" or "1" to force rebuilding the bootstrap image
# PUSH_IMAGES: "true" or "1" to push images to the pinterest elastic container registry
# SKIP_BUILD: "true" or "1" to skip building the base && vitess docker images with bootstrap + local changes
# TARBALL_GZ: Output location of the build artifact
# GIT_COMMIT: Label to use for docker images and the build artifact
# PACKAGE_DEB: "true" or "1" to build a .deb file and upload it
# BUILD_DIR: Directory where the built .deb file should be placed.

REGISTRY="998131032990.dkr.ecr.us-east-1.amazonaws.com"

if [ "$REBUILD_COMMON" == 'true' ] || [ "$REBUILD_COMMON" == '1' ]
then
  ./docker/bootstrap/build.sh common
  ./docker/bootstrap/build.sh percona

  docker tag vitess/bootstrap:percona $REGISTRY/vitess/bootstrap:latest
  docker tag vitess/bootstrap:percona $REGISTRY/vitess/bootstrap:percona

  if [ "$PUSH_IMAGES" == 'true' ] || [ "$PUSH_IMAGES" == '1' ]
  then
    docker push $REGISTRY/vitess/bootstrap:latest
    docker push $REGISTRY/vitess/bootstrap:percona
  fi
fi

# SKIP_BUILD can be set to use pre-built images from the repo at the provided git commit version
if [ "$SKIP_BUILD" != 'true' ] && [ "$PUSH_IMAGES" != '1' ]
then
  # Copy the most recent files into the bootstrap image to create a base image
  docker build --no-cache -f docker/base/Dockerfile.percona --build-arg BASE_IMAGE=$REGISTRY/vitess/bootstrap:percona -t $REGISTRY/vitess/base:"$GIT_COMMIT" .

  # Run unit tests and build a pinterest-specific image
  docker build --no-cache -f Dockerfile.pinterest --build-arg BASE_IMAGE=$REGISTRY/vitess/base:"$GIT_COMMIT" -t $REGISTRY/vitess:"$GIT_COMMIT" .

  # Build vtctld specific image
  docker build --no-cache -f Dockerfile.vtctld.pinterest --build-arg BASE_IMAGE=$REGISTRY/vitess/base:"$GIT_COMMIT" -t $REGISTRY/vitess/vtctld:"$GIT_COMMIT" .
fi

# Unit tests pass, making the build artifact succeeded. Let's push the base and vtgate images out!
if [ "$PUSH_IMAGES" == 'true' ] || [ "$PUSH_IMAGES" == '1' ]
then
  docker push $REGISTRY/vitess/base:"$GIT_COMMIT"
  docker push $REGISTRY/vitess:"$GIT_COMMIT"
  docker push $REGISTRY/vitess/vtctld:"$GIT_COMMIT"

  # Tag $REGISTRY/vitess:latest for easier access
  docker tag $REGISTRY/vitess:"$GIT_COMMIT" $REGISTRY/vitess:latest
  docker push $REGISTRY/vitess:latest
fi

if [ "$PACKAGE_DEB" == 'true' ] || [ "$PACKAGE_DEB" == '1' ]
then
  docker run -i $REGISTRY/vitess:"$GIT_COMMIT" /vt/scripts/write_build_artifact_to_stdout.sh > "$TARBALL_GZ"
# Package a .deb file where /vt/* has the contents of the build artifact
  DATE=$(date +%Y%m%d.%H%M)

  OUTPUT=$(fpm --verbose \
    -s tar \
    -t deb \
    -p "$BUILD_DIR" \
    -n "vitess" \
    -v "$DATE" \
    -a amd64 \
    --category "Util" \
    --vendor "Pinterest" \
    --deb-no-default-config-files \
    -m "Pinterest Ops" \
    --prefix /vt/ \
    "$TARBALL_GZ")
  EXIT="${?}"

  if [[ "${EXIT}" -ne 0 ]]; then
    echo -e "ERROR building package\n${OUTPUT}"
    exit 255
  fi

  #deb-s3 upload --preserve-versions --visibility private --fail-if-exists --bucket pinterest-repo-trusty/apt --arch amd64 --sign BC0BEAD1 --codename trusty $BUILD_DIR/*.deb
  #vitess_20190410.1534_all.deb
  DEB_PATH=$(find "$BUILD_DIR" -name '*.deb' -print -quit)

  if [[ ! -f "$DEB_PATH" ]]; then
    echo -e "ERROR locating deb file at:${DEB_PATH}"
    exit 255
  fi

  cat << EOF > "${WORKSPACE}/vitess.deb.properties"
ARTIFACT_PATH=${DEB_PATH}
FROM=JOB
FROM_PROJECT=${JOB_NAME}
FROM_BUILD=${BUILD_NUMBER}
ENABLE_UPLOAD_METADATA=TRUE
ENABLE_UPLOAD_METADATA_DEB=TRUE
ARTIFACTORY_SERVER_ENV=prod
ARTIFACTORY_SERVER_REGION=use1
ARTIFACTORY_DESTINATION_REPO=ubuntu-deb-apt-prod-local
TO_REPO=ubuntu-deb-apt-prod-local
ENABLE_UPLOAD_METADATA_DEB_S3=TRUE
UPLOAD_METADATA_DEB_DISTROS=trusty,xenial,bionic
UPLOAD_METADATA_DEB_COMPONENT=main
UPLOAD_METADATA_DEB_ARCHITECTURE=amd64
EOF

fi
