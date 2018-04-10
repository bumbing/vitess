# Tarball file should be the output file from "bazel build :full_dist"
TARBALL="$1"

# Output will go in the debs/ directory.
mkdir -p debs/

DATE=$(date +%Y%m%d.%H%M)

fpm --verbose \
  -s tar \
  -t deb \
  -p debs/ \
  -n "vitess" \
  -v $DATE \
  -a all \
  --category "Util" \
  --vendor "Pinterest" \
  --deb-no-default-config-files \
  -m "Pinterest Ops" \
  --prefix /vt/build/ \
  "$TARBALL"
