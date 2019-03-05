#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

TELETRAAN_ZK_SERVERS=vitess-infra-zookeeper-dev-001:2181,vitess-infra-zookeeper-dev-002:2181,vitess-infra-zookeeper-dev-003:2181,vitess-infra-zookeeper-dev-004:2181,vitess-infra-zookeeper-dev-005:2181 \
    ./scripts/vtgate_startup.sh -tablet_types_to_wait="" "$@"
