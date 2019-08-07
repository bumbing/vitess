#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# The upgrade of cloud.google.com/go is to avoid an incompatibility with the version of gax-go
# that "go mod init" picks.

# The test command compiles tests without running any.

go mod init && \
    sed -i '' 's_cloud.google.com/go .*_cloud.google.com/go v0.36.0_' go.mod
