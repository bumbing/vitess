#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

go run test.go -docker=false -print-log "$@"
