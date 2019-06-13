#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Pass -d as the first argument to add flags for docker (mostly related to log location)
#
# Any command line arguments not parsed here will be passed on to vtgate. You can think of
# the values here as defaults, since if you do "vtgate -flag value1 -flag value2" the second
# setting of -flag will override the first, setting -flag to value2.

args=(
  -topo_implementation zk2
  -topo_global_server_address "${TELETRAAN_ZK_SERVERS}"
  -topo_global_root "${TELETRAAN_TOPO_GLOBAL_ROOT:-/vitess/global}"
  -port 15001
  -grpc_port 15991
  -mysql_server_port 3306
  -mysql_tcp_version tcp4
  -mysql_auth_server_impl knox
  -knox_supported_roles "scriptro,longqueryro,scriptrw,longqueryrw${TELETRAAN_ADDITIONAL_KNOX_ROLES:+,}${TELETRAAN_ADDITIONAL_KNOX_ROLES:-}"
  -knox_role_mapping "scriptro:reader,longqueryro:reader,pepsirw:reader:writer:admin,devpepsirw:reader:writer:admin,scriptrw:reader:writer:admin,pepsilong:reader:writer:admin,devpepsilong:reader:writer:admin,longqueryrw:reader:writer:admin"
  -grpc_keepalive_time 30s
  -cell "${TELETRAAN_CELL:-test}"
  -cells_to_watch "${TELETRAAN_CELL:-test}"
  -tablet_types_to_wait "${TELETRAAN_TABLET_TYPES_TO_WAIT:-MASTER,REPLICA}"
  -gateway_implementation discoverygateway
  # If at least one rdonly instance has lag <10m and the others do not,
  # prefer the one with <10m lag. Indexing aims to have data served within
  # 15m of being written to patio, so we have a more extreme prefernce for
  # non-lagging rdonly replicas than the discovery gateway default.
  # The default is to avoid replicas with >2h of lag if at least two other
  # replicas have less lag.
  -discovery_high_replication_lag_minimum_serving 10m
  -enable_buffer_dry_run
  -min_number_serving_vttablets 1
  -service_map 'grpc-vtgateservice'
  "-merge_keyspace_joins_to_single_shard=${TELETRAAN_MERGE_JOIN_SHARDS:-true}"
  -allow_select_unauthoritative_col
  -alsologtostderr
  -mysql_server_query_timeout 2h
  -mysql_user_query_timeouts "scriptro:10s,scriptrw:10s,pepsirw:10s,devpepsirw:10s"
  -mysql_server_ssl_cert /var/lib/normandie/fuse/cert/generic
  -mysql_server_ssl_key /var/lib/normandie/fuse/key/generic
  -mysql_server_ssl_ca /var/lib/normandie/fuse/ca/generic
  -mysql_server_ssl_reload_frequency 15m
  -emit_stats
  -stats_emit_period 1m
  -stats_backend opentsdb
  "-opentsdb_service=${TELETRAAN_TSDB_SERVICE:-}"
  "-discourage-v2-inserts=${TELETRAAN_DISCOURAGE_V2_INSERTS:-false}"
  "-merge_left_join_unique_vindexes=${TELETRAAN_MERGE_LEFT_JOINS:-false}"
)

if [ "${VTGATE_COMMAND:-}" = "" ]
then
   VTGATE_COMMAND="go run vitess.io/vitess/go/cmd/vtgate "
fi

while getopts ":d" opt; do
  case $opt in
    d)
      # Running in docker
      args+=(
        -log_dir /vt/logs
        -log_queries_to_file /vt/logs/queries.log
        -pid_file /vt/vtdataroot/vtgate.pid
      )
      VTGATE_COMMAND="/vt/bin/vtgate"
      ;;
    \?)
      break
      ;;
  esac
done
shift $((OPTIND -1))


if [[ ! -z "${TELETRAAN_ALLOWED_TABLET_TYPES:-}" ]]; then
  args+=(
    -allowed_tablet_types "${TELETRAAN_ALLOWED_TABLET_TYPES}"
  )
fi

if [[ "${TELETRAAN_ENFORCE_TLS_HOST:-}" == "dev" ]]; then
  args+=(
    -group_tls_regexes "writer:.*(pepsi|patio|cola|dev-|mysql-open-access-bastion).*"
  )
elif [[ "${TELETRAAN_ENFORCE_TLS_HOST:-}" == "prod" ]]; then
  args+=(
    -group_tls_regexes "writer:^(m10n-pepsi-(prod|canary|long-jobs-prod|cron|canary)-.*)|^cloudeng-sox-mysql-open-access-bastion-prod-.*"
  )
fi

if [[ "${TELETRAAN_DISABLE_TLS:-false}" == "true" ]]; then
  args+=(
    "-mysql_server_ssl_cert="
    "-mysql_server_ssl_key="
  )
fi

if [[ ! -z "${TELETRAAN_DARK_MAX_ROWS:-}" ]]; then
  args+=(
    -pinterest_dark_read_max_compared_rows "${TELETRAAN_DARK_MAX_ROWS}"
  )
fi

if [[ ! -z "${TELETRAAN_DARK_LIGHT_TARGET:-}" ]]; then 
  args+=(
    -pinterest_dark_read_light_target "${TELETRAAN_DARK_LIGHT_TARGET}"
  )
fi

${VTGATE_COMMAND:-/vt/bin/vtgate} "${args[@]}" "$@"
