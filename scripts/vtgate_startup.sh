#!/bin/bash

# Pass -d as the first argument to add flags for docker (mostly related to log location)
#
# Any command line arguments not parsed here will be passed on to vtgate. You can think of
# the values here as defaults, since if you do "vtgate -flag value1 -flag value2" the second
# setting of -flag will override the first, setting -flag to value2.

DOCKER=false
DEV=false
LATEST=false

EXTRA_ARGS=""
if [ "$VTGATE_COMMAND" = "" ]
then
   VTGATE_COMMAND="bazel run --workspace_status_command=./scripts/workspace_status.sh go/cmd/vtgate --"
fi

while getopts ":det" opt; do
  case $opt in
    d)
      DOCKER=true
      ;;
    e)
      DEV=true
      ;;
    t)
      LATEST=true
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
  VTGATE_COMMAND="/vt/bin/vtgate"
fi

# For new command line arguments that may be enabled for dev but not prod (yet).
if [[ ${LATEST} == true ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -opentsdb_service vtgate_latest"
fi

# For new command line arguments that may be enabled for dev but not prod (yet).
if [[ ${DEV} == true ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -emit_stats=false \
    -opentsdb_service vtgate_test"
fi

if [[ ! -z "${TELETRAAN_ZK_SERVERS}" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -topo_global_server_address ${TELETRAAN_ZK_SERVERS}"
fi

if [[ ! -z "${TELETRAAN_TABLET_TYPES_TO_WAIT}" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -tablet_types_to_wait ${TELETRAAN_TABLET_TYPES_TO_WAIT}"
fi

if [[ ! -z "${TELETRAAN_TOPO_GLOBAL_ROOT}" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -topo_global_root ${TELETRAAN_TOPO_GLOBAL_ROOT}"
fi

if [[ ! -z "${TELETRAAN_CELL}" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -cell ${TELETRAAN_CELL} \
    -cells_to_watch ${TELETRAAN_CELL}"
fi

if [[ ! -z "${TELETRAAN_ALLOWED_TABLET_TYPES}" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -allowed_tablet_types ${TELETRAAN_ALLOWED_TABLET_TYPES}"
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
  -grpc_keepalive_time 30s \
  -cell test \
  -cells_to_watch test \
  -tablet_types_to_wait MASTER,REPLICA \
  -gateway_implementation discoverygateway \
  -service_map 'grpc-vtgateservice' \
  -opentsdb_service vtgate \
  -emit_stats \
  -stats_emit_period 1m \
  -stats_backend opentsdb \
  -merge_keyspace_joins_to_single_shard \
  -alsologtostderr \
  -opentsdb_service vtgate \
  -mysql_server_query_timeout 2h \
  -mysql_user_query_timeouts scriptro:10s,scriptrw:10s \
  ${EXTRA_ARGS} \
  $@

