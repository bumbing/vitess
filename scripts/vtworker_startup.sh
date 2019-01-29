#!/bin/bash

# Pass -d as the first argument to add flags for docker (mostly related to log location)
#
# Any command line arguments not parsed here will be passed on to vtworker. You can think of
# the values here as defaults, since if you do "vtworker -flag value1 -flag value2" the second
# setting of -flag will override the first, setting -flag to value2.

DOCKER=false
LATEST=false

EXTRA_ARGS=""
if [ "$VTWORKER_COMMAND" = "" ]
then
   VTWORKER_COMMAND="go run vitess.io/vitess/go/cmd/vtworker "
fi

while getopts ":det" opt; do
  case $opt in
    d)
      DOCKER=true
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
    -pid_file /tmp/vtworker.pid"
  VTWORKER_COMMAND="/vt/bin/vtworker"
fi

if [[ ${LATEST} == true || "$STAGE_NAME" == "shadow" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -security_policy= "
fi

if [[ ! -z "${TELETRAAN_TSDB_SERVICE}" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -emit_stats \
    -stats_emit_period 1m \
    -stats_backend opentsdb \
    -opentsdb_service ${TELETRAAN_TSDB_SERVICE}"
fi

if [[ ! -z "${TELETRAAN_ZK_SERVERS}" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -topo_global_server_address ${TELETRAAN_ZK_SERVERS}"
fi

if [[ ! -z "${TELETRAAN_TOPO_GLOBAL_ROOT}" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -topo_global_root ${TELETRAAN_TOPO_GLOBAL_ROOT}"
fi

if [[ ! -z "${TELETRAAN_CELL}" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -cell ${TELETRAAN_CELL}"
fi

${VTWORKER_COMMAND} \
  -topo_implementation zk2 \
  -topo_global_root /vitess/global \
  -port 15001 \
  -grpc_port 15991 \
  -cell test \
  -service_map 'grpc-vtworker' \
  -security_policy role_whitelist \
  -whitelisted_roles monitoring,debugging \
  -username vtworker \
  -groups reader,writer,admin \
  -alsologtostderr \
  -use_v3_resharding_mode \
  ${EXTRA_ARGS} \
  $@

