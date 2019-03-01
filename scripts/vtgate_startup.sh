#!/bin/bash

# Pass -d as the first argument to add flags for docker (mostly related to log location)
#
# Any command line arguments not parsed here will be passed on to vtgate. You can think of
# the values here as defaults, since if you do "vtgate -flag value1 -flag value2" the second
# setting of -flag will override the first, setting -flag to value2.

DOCKER=false

EXTRA_ARGS=""
if [ "$VTGATE_COMMAND" = "" ]
then
   VTGATE_COMMAND="go run vitess.io/vitess/go/cmd/vtgate "
fi

while getopts ":det" opt; do
  case $opt in
    d)
      DOCKER=true
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

if [[ "${TELETRAAN_ENFORCE_TLS_HOST}" == "dev" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -group_tls_regexes writer:.*(pepsi|patio|cola|dev-|mysql-open-access-bastion).*"
fi

if [[ "${TELETRAAN_ENFORCE_TLS_HOST}" == "prod" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -group_tls_regexes writer:^(m10n-(pepsi|patio|croncola)(-long-jobs)?-(prod|cron|canary|staging)-.*)|cloudeng-mysql-open-access-bastion-prod-.*"
fi

if [[ "${TELETRAAN_DISABLE_TLS}" == "true" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -mysql_server_ssl_cert= \
    -mysql_server_ssl_key= "
fi

if [[ "${TELETRAAN_DARK_GATE}" == "true" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -pinterest_dark_read_gate "
fi

if [[ ! -z "${TELETRAAN_DARK_MAX_ROWS}" ]]; then
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -pinterest_dark_read_max_compared_rows ${TELETRAAN_DARK_MAX_ROWS} "
fi

if [[ ! -z "${TELETRAAN_DARK_LIGHT_TARGET}" ]]; then 
  EXTRA_ARGS=" \
    ${EXTRA_ARGS} \
    -pinterest_dark_read_light_target ${TELETRAAN_DARK_LIGHT_TARGET} "
fi

SUPPORTED_KNOX_ROLES="scriptro,longqueryro,scriptrw,longqueryrw"
if [[ ! -z "${TELETRAAN_ADDITIONAL_KNOX_ROLES}" ]]; then
  SUPPORTED_KNOX_ROLES="${SUPPORTED_KNOX_ROLES},${TELETRAAN_ADDITIONAL_KNOX_ROLES}"
fi

# TODO(dweitzman): To require TLS for writing, we'll do something like this:
# -group_tls_regexes "writer:^m10n-pepsi-prod..*,admin:^m10n-pepsi-prod..*"
# To test with a devapp, the regex might look more like this:
#  -group_tls_regexes "writer:^dev-dweitzman\\..*"
${VTGATE_COMMAND} \
  -topo_implementation zk2 \
  -topo_global_root /vitess/global \
  -port 15001 \
  -grpc_port 15991 \
  -mysql_server_port 3306 \
  -mysql_tcp_version tcp4 \
  -mysql_server_socket_path /tmp/mysql.sock \
  -mysql_auth_server_impl knox \
  -knox_supported_roles "${SUPPORTED_KNOX_ROLES}" \
  -knox_role_mapping scriptro:reader,longqueryro:reader,pepsirw:reader:writer:admin,devpepsirw:reader:writer:admin,scriptrw:reader:writer:admin,pepsilong:reader:writer:admin,devpepsilong:reader:writer:admin,longqueryrw:reader:writer:admin \
  -grpc_keepalive_time 30s \
  -cell test \
  -cells_to_watch test \
  -tablet_types_to_wait MASTER,REPLICA \
  -gateway_implementation discoverygateway \
  -service_map 'grpc-vtgateservice' \
  -merge_keyspace_joins_to_single_shard \
  -allow_select_unauthoritative_col \
  -alsologtostderr \
  -mysql_server_query_timeout 2h \
  -mysql_user_query_timeouts scriptro:10s,scriptrw:10s,pepsirw:10s,devpepsirw:10s \
  -mysql_server_ssl_cert /var/lib/normandie/fuse/cert/generic \
  -mysql_server_ssl_key /var/lib/normandie/fuse/key/generic \
  -mysql_server_ssl_ca /var/lib/normandie/fuse/ca/generic \
  -mysql_server_ssl_reload_frequency 15m \
  ${EXTRA_ARGS} \
  $@

