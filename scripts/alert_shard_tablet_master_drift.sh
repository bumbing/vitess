#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

if command -v /usr/local/bin/vitess_utils/vitess_utils.sh >/dev/null; then
    /usr/local/bin/vitess_utils/vitess_utils.sh --action validateshardtabletmaster
fi
