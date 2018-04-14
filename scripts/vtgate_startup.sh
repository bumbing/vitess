#!/bin/bash

# Pass -d as the first argument to add flags for docker (mostly related to log location)
#
# Any command line arguments not parsed here will be passed on to vtgate. You can think of
# the values here as defaults, since if you do "vtgate -flag value1 -flag value2" the second
# setting of -flag will override the first, setting -flag to value2.

DOCKER=false
DEV=false

EXTRA_ARGS=""
if [ "$VTGATE_COMMAND" = "" ]
then
   VTGATE_COMMAND="bazel run --workspace_status_command=./scripts/workspace_status.sh go/cmd/vtgate --"
fi

while getopts ":de" opt; do
  case $opt in
    d)
      DOCKER=true
      ;;
    e)
      DEV=true
      ;;
    \?)
      break
      ;;
  esac
done
shift $((OPTIND -1))

if [[ ${DOCKER} == true ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -log_dir /vt/logs \
    -log_queries_to_file /vt/logs/queries.log \
    -pid_file /vt/vtdataroot/vtgate.pid"
  VTGATE_COMMAND="/vt/build/bin/vtgate"
fi

# For new command line arguments that may be enabled for dev but not prod (yet).
if [[ ${DEV} == true ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -emit_stats=false \
    -opentsdb_service vtgate_test"
fi

${VTGATE_COMMAND} \
  -topo_implementation zk2 \
  -topo_global_root /vitess/global \
  -port 15001 \
  -grpc_port 15991 \
  -mysql_server_port 3306 \
  -mysql_tcp_version tcp4 \
  -mysql_server_socket_path /tmp/mysql.sock \
  -mysql_auth_server_impl knox \
  -knox_supported_roles scriptro,longqueryro,scriptrw,longqueryrw \
  -cell test \
  -cells_to_watch test \
  -tablet_types_to_wait MASTER,REPLICA \
  -gateway_implementation discoverygateway \
  -service_map 'grpc-vtgateservice' \
  -opentsdb_service vtgate \
  -emit_stats \
  -stats_emit_period 1m \
  -stats_backend opentsdb \
  -alsologtostderr \
  -opentsdb_service vtgate \
  ${EXTRA_ARGS} \
  $@

