#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ZK_ADDRS=vitess-infra-zookeeper-dev-001:2181,vitess-infra-zookeeper-dev-002:2181,vitess-infra-zookeeper-dev-003:2181,vitess-infra-zookeeper-dev-004:2181,vitess-infra-zookeeper-dev-005:2181

./scripts/vtgate_startup.sh -e -topo_global_server_address $ZK_ADDRS -tablet_types_to_wait="" $@
