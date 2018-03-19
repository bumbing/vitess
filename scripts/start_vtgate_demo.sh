#!/bin/bash
BAZEL_RUN="bazel run"

if [ -x "$(command -v ibazel)" ]; then
	BAZEL_RUN="ibazel run"
fi

ZK_ADDRS=vitess-infra-zookeeper-dev-001:2181,vitess-infra-zookeeper-dev-002:2181,vitess-infra-zookeeper-dev-003:2181,vitess-infra-zookeeper-dev-004:2181,vitess-infra-zookeeper-dev-005:2181

$BAZEL_RUN go/cmd/vtgate -- -logtostderr -topo_implementation zk2 -topo_global_server_address $ZK_ADDRS -port 15001 -cell test -topo_global_root /vitess/global -cells_to_watch test -gateway_implementation discoverygateway -service_map grpc-vtgateservice -pid_file /tmp/vtgate.pid -mysql_server_port 15306 -mysql_auth_server_impl none -knox_supported_usernames scriptrw,longqueryrw,scriptro,longqueryro $@
