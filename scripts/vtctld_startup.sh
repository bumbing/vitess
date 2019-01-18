#!/bin/bash

# Pass -d as the first argument to add flags for docker (mostly related to log location)
#
# Any command line arguments not parsed here will be passed on to vtctld. You can think of
# the values here as defaults, since if you do "vtctld -flag value1 -flag value2" the second
# setting of -flag will override the first, setting -flag to value2.

DOCKER=false
LATEST=false

EXTRA_ARGS=""
if [ "$VTCTLD_COMMAND" = "" ]
then
   VTCTLD_COMMAND="go run vitess.io/vitess/go/cmd/vtctld "
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
    -pid_file /tmp/vtctld.pid"
  VTCTLD_COMMAND="/vt/bin/vtctld"
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

${VTCTLD_COMMAND} \
  -topo_implementation zk2 \
  -topo_global_root /vitess/global \
  -port 15001 \
  -grpc_port 15991 \
  -cell test \
  -service_map 'grpc-vtctl' \
  -security_policy role_whitelist \
  -whitelisted_roles monitoring,debugging \
  -alsologtostderr \
  -web_dir /vt/web/vtctld \
  -web_dir2 /vt/web/vtctld2/app \
  -workflow_manager_init \
  -workflow_manager_use_election \
  -backup_storage_implementation file \
  -file_backup_storage_root $VTDATAROOT/backups \
  ${EXTRA_ARGS} \
  $@

