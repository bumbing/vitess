# Tarball file should be the output file from "bazel build :full_dist"
# Usage:
# ./scripts/package_deb.sh <tarball_name> <output dir for .deb file>
OUTPUT_DIR="$2"
TARBALL="$1"

DATE=$(date +%Y%m%d.%H%M)

fpm --verbose \
  -s tar \
  -t deb \
  -p "$OUTPUT_DIR" \
  -n "vitess" \
  -v $DATE \
  -a all \
  --category "Util" \
  --vendor "Pinterest" \
  --deb-no-default-config-files \
  -m "Pinterest Ops" \
  --prefix /vt/build/ \
  "$TARBALL"
