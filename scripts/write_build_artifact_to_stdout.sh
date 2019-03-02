#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

cd /vt/ && (tar cf - ./ | gzip -f -)
