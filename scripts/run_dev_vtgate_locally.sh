#!/bin/bash
if [ -x "$(command -v ibazel)" ]; then
	VTGATE_COMMAND="ibazel run go/cmd/vtgate --"
fi

ZK_ADDRS=vitess-infra-zookeeper-dev-001:2181,vitess-infra-zookeeper-dev-002:2181,vitess-infra-zookeeper-dev-003:2181,vitess-infra-zookeeper-dev-004:2181,vitess-infra-zookeeper-dev-005:2181

VTGATE_COMMAND="$VTGATE_COMMAND" ./scripts/vtgate_startup.sh -e -topo_global_server_address $ZK_ADDRS $@
