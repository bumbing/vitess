#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Pass -d as the first argument to add flags for docker (mostly related to log location)
#
# Any command line arguments not parsed here will be passed on to vtctld. You can think of
# the values here as defaults, since if you do "vtctld -flag value1 -flag value2" the second
# setting of -flag will override the first, setting -flag to value2.

args=(
  -topo_implementation zk2
  -topo_global_root "${TELETRAAN_TOPO_GLOBAL_ROOT:-/vitess/global}"
  -topo_global_server_address "${TELETRAAN_ZK_SERVERS}"
  -port 15001
  -grpc_port 15991
  -cell "${TELETRAAN_CELL:-test}"
  -service_map 'grpc-vtctl'
  -security_policy role_whitelist
  -whitelisted_roles monitoring,debugging
  -alsologtostderr
  -web_dir /vt/web/vtctld
  -web_dir2 /vt/web/vtctld2/app
  -workflow_manager_init
  -workflow_manager_use_election
)

if [[ "${COMPILE:-}" ]]; then
  VTCTLD_COMMAND="go run vitess.io/vitess/go/cmd/vtctld"
fi

while getopts ":dt" opt; do
  case $opt in
    d)
      # Running in docker. Log to known locations.
      args+=(
        -log_dir /vt/logs
        -pid_file /tmp/vtctld.pid
      );;
    t)
      # Relax security policy in dev and allow queries
      args+=(
        "-security_policy="
        -enable_queries
      );;
    \?)
      break
      ;;
  esac
done
shift $((OPTIND -1))

# Relax security policy in shadow so we can view all the status.
if [[ "${STAGE_NAME:-}" == "shadow" ]]; then
  args+=(
    "-security_policy="
  )
fi

if [[ "${TELETRAAN_TSDB_SERVICE:-}" ]]; then
  args+=(
    -emit_stats
    -stats_emit_period 1m
    -stats_backend opentsdb
    -opentsdb_service "${TELETRAAN_TSDB_SERVICE}"
  )
fi

${VTCTLD_COMMAND:-/vt/bin/vtctld} "${args[@]}" "$@"
