# Docker tag for prod probably looks like 998131032990.dkr.ecr.us-east-1.amazonaws.com/vitess:$GIT_COMMIT
# Build artifact file on jenkins probably looks like `pwd`/BUILD_DIR/vitess-${GIT_COMMIT:0:7}.tar.gz

# Usage: ./scripts/build_artifact_and_prod_docker_image.sh <build_artifact_file> <docker tag>

set -e

if [[ "$1" = "" || "$2" = "" || "$3" != "" ]]
then
  echo "Usage: ./scripts/build_artifact_and_prod_docker_image.sh <build_artifact_file> <docker tag>"
  exit 1
fi

FULL_OUTPUT_FILE="$1"

if [[ "${FULL_OUTPUT_FILE:0:1}" != "/" ]]
then
  FULL_OUTPUT_FILE="`pwd`/$FULL_OUTPUT_FILE"
fi

PROD_IMAGE_NAME="$2"

OUTPUT_LOCAL_DIR="$(dirname $FULL_OUTPUT_FILE)"
OUTPUT_TARBALL_NAME="$(basename $FULL_OUTPUT_FILE)"

mkdir -p $OUTPUT_LOCAL_DIR

echo "Building :full_dist as $FULL_OUTPUT_FILE using a dockerized version of bazel..."

docker run -t -v `pwd`:/vt/src -v $OUTPUT_LOCAL_DIR:/vt/build_artifacts -w /vt/src --entrypoint /bin/bash 998131032990.dkr.ecr.us-east-1.amazonaws.com/bazel:latest -c "./bazel_bootstrap.sh && bazel build --symlink_prefix=/ :full_dist && cp -f \`bazel info bazel-bin\`/full_dist.tar.gz /vt/build_artifacts/$OUTPUT_TARBALL_NAME"


echo Finished building :full_dist as $FULL_OUTPUT_FILE
echo Copying the output into a docker image and tagging as $PROD_IMAGE_NAME...

# TODO(dweitzman): Improve the dockerfile story so we don't have to write it to disk.
# Piping to stdin seemed to work locally, but there were issues on jenkins and technically
# I think docker claims that if you use "-" as a filename it doesn't support reading files
# from your working directory.

DOCKERFILE_NAME="$OUTPUT_LOCAL_DIR/vtgate.dockerfile.tmp"

cat <<EOF > "$DOCKERFILE_NAME"
FROM 998131032990.dkr.ecr.us-east-1.amazonaws.com/ubuntu14.04:latest
ADD $OUTPUT_TARBALL_NAME /vt/build/
EOF

docker build --pull --no-cache -t "$PROD_IMAGE_NAME" -f "$DOCKERFILE_NAME" "$OUTPUT_LOCAL_DIR"
rm "$DOCKERFILE_NAME"

echo Finished tagging docker image $PROD_IMAGE_NAME
