ENV_NAME="$1"

if [[ "$ENV_NAME" == "" ]]; then
   echo "Usage: pvtcl.sh [vitess env] [args to vtctl...]"
   exit 1
fi

ENV_EXISTS=$(cat /var/config/config.services.vitess_environments_config | jq -r ".${ENV_NAME} | select (.!=null) ")

if [[ "$ENV_EXISTS" == "" ]]; then
   echo "Looks like an invalid env name. Valid names tend to look like prod, latest, test, etc."
   echo "Usage: pvtcl.sh [vitess env] [args to vtctl...]"
   exit 1
fi

# devapps have jq 1.3 which doesn't yet support "join", so xargs is used to join the host list
ZK_SERVERS=$(cat /var/config/config.services.vitess_environments_config | jq -r ".${ENV_NAME}.zk.hosts[]" | xargs  | sed -e 's/ /,/g')
ZK_ROOT=$(cat /var/config/config.services.vitess_environments_config | jq -r ".${ENV_NAME}.zk.global_root")

shift

set -e

# TODO(dweitzman): Ideally we'd point at vtctld here, but we don't have fixed hostnames for vtctld at this time.

/vt/bin/vtctl -topo_implementation zk2 -topo_global_server_address $ZK_SERVERS -topo_global_root $ZK_ROOT "$@"

