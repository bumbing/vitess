#!/bin/bash

# Pass -d as the first argument to add flags for docker (mostly related to log location)
#
# Any command line arguments not parsed here will be passed on to vtctld. You can think of
# the values here as defaults, since if you do "vtctld -flag value1 -flag value2" the second
# setting of -flag will override the first, setting -flag to value2.

DOCKER=false
DEV=false
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
    -pid_file /vt/vtdataroot/tmp/vtctld.pid"
  VTCTLD_COMMAND="/vt/bin/vtctld"
fi

# For new command line arguments that may be enabled for dev but not prod (yet).
if [[ ${LATEST} == true ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -security_policy= \
    -opentsdb_service vtctld_latest"
fi

# For new command line arguments that may be enabled for dev but not prod (yet).
if [[ ${DEV} == true ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -emit_stats=false \
    -opentsdb_service vtctld_test"
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
  -port 15000 \
  -grpc_port 15999 \
  -cell test \
  -service_map 'grpc-vtctl' \
  -security_policy role_whitelist \
  -whitelisted_roles monitoring,debugging \
  -opentsdb_service vtctld \
  -emit_stats \
  -stats_emit_period 1m \
  -stats_backend opentsdb \
  -alsologtostderr \
  -web_dir $VTTOP/web/vtctld \
  -web_dir2 $VTTOP/web/vtctld2/app \
  -workflow_manager_init \
  -workflow_manager_use_election \
  -backup_storage_implementation file \
  -file_backup_storage_root $VTDATAROOT/backups \
  ${EXTRA_ARGS} \
  $@ &

